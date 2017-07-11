package rabbitmq

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"crypto/tls"

	"github.com/streadway/amqp"
	"github.com/upfluence/goutils/log"
)

// AMQPChannel is an interface over amqp.Channel
type AMQPChannel interface {
	Consume(string, string, bool, bool, bool, bool, amqp.Table) (<-chan amqp.Delivery, error)
	ExchangeDeclare(string, string, bool, bool, bool, bool, amqp.Table) error
	NotifyClose(chan *amqp.Error) chan *amqp.Error
	Publish(string, string, bool, bool, amqp.Publishing) error
	Qos(int, int, bool) error
	QueueBind(string, string, string, bool, amqp.Table) error
	QueueDeclare(string, bool, bool, bool, bool, amqp.Table) (amqp.Queue, error)
}

// AMQPConnection is an interface over amqp.Connection
type AMQPConnection interface {
	Channel() (AMQPChannel, error)
	Close() error
}

// Connection is a wrapper for amqp.Connection needed to be able
// to assign the result of the dialer function to
// RabbitMQTransport.Connection, because Go doesn't support
// covariant return types
// Relevant discussion: https://github.com/streadway/amqp/issues/164
type Connection struct {
	*amqp.Connection
}

// Channel is a wrapper for amqp.Connection.Channel()
func (c *Connection) Channel() (AMQPChannel, error) {
	return c.Connection.Channel()
}

// Close is a wrapper for amqp.Connection.Close()
func (c *Connection) Close() error {
	return c.Connection.Close()
}

// RabbitMQTransport contains AMQP objects required to communicate with RabbitMQ
type RabbitMQTransport struct {
	Connection     AMQPConnection
	Channel        AMQPChannel
	ClosingChannel chan bool
	Configs        []*TransportConfig
	dialer         func(string) (AMQPConnection, error)
	dialerConfig   func(string, amqp.Config) (AMQPConnection, error)
}

// NewRabbitMQTransport creates a RabbitMQTransport instance from a given URI
func NewRabbitMQTransport(uri string) (*RabbitMQTransport, error) {
	config, err := NewTransportConfig(uri)

	if err != nil {
		return nil, fmt.Errorf("Received invalid URI: %s", err)
	}

	return NewRabbitMQHATransport([]*TransportConfig{config}), nil
}

func amqpDialer(url string) (AMQPConnection, error) {
	var conn = &Connection{}
	var err error
	conn.Connection, err = amqp.Dial(url)

	return conn, err
}

func amqpDialerConfig(url string, config amqp.Config) (AMQPConnection, error) {
	var conn = &Connection{}
	var err error
	conn.Connection, err = amqp.DialConfig(url, config)

	return conn, err
}

// NewRabbitMQHATransport creates a RabbitMQTransport instance from a list of
// TransportConfig objects in order to connect to a
// High Availability RabbitMQ cluster
func NewRabbitMQHATransport(configs []*TransportConfig) *RabbitMQTransport {
	return &RabbitMQTransport{
		ClosingChannel: make(chan bool),
		Configs:        configs,
		dialer:         amqpDialer,
		dialerConfig:   amqpDialerConfig,
	}
}

func (t *RabbitMQTransport) GetClosingChan() chan bool {
	return t.ClosingChannel
}

func (t *RabbitMQTransport) Connect() error {
	var (
		uri           string
		err           error
		randGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
	)

	var config *TransportConfig
	for _, idx := range randGenerator.Perm(len(t.Configs)) {
		config = t.Configs[idx]
		uri = config.GetURI()

		log.Noticef("Trying to connect to URI: %s", uri)

		c := amqp.Config{
			Heartbeat: 10 * time.Second, // amqp.defaultHeartbeat
		}

		if config.Heartbeat.String()!= "" {
			c.Heartbeat, err = time.ParseDuration(config.Heartbeat.String() + "s")

			if err != nil {
				log.Warningf(
					"Failed to parse the heartbeat value \"%s\": %s",
					c.Heartbeat,
					err.Error(),
				)
				continue
			}
		}

		if config.Ssl.CertChainFile != "" && config.Ssl.PrivateKeyFile != "" {

			certs, err := tls.LoadX509KeyPair(config.Ssl.CertChainFile, config.Ssl.PrivateKeyFile)
			if err != nil {
				log.Errorf("Failed to load ssl key/cert: %s", err.Error())
				continue
			}

			// Skip tls verify as non matching cert common names are valid in sensu proper.
			// Perhaps this should be a config option.
			c.TLSClientConfig = &tls.Config{
				Certificates: []tls.Certificate{certs},
				InsecureSkipVerify: true,
			}
		}

		t.Connection, err = t.dialerConfig(uri, c)

		if err != nil {
			log.Warningf("Failed to connect to URI \"%s\": %s", uri, err.Error())
			continue
		}

		break
	}

	if err != nil {
		log.Errorf("RabbitMQ connection error: %s", err.Error())
		return err
	}

	t.Channel, err = t.Connection.Channel()

	if err != nil {
		log.Errorf("RabbitMQ channel error: %s", err.Error())
		return err
	}

	if prefetchString := config.Prefetch.String(); prefetchString != "" {
		var prefetch int
		prefetch, err = strconv.Atoi(prefetchString)

		if err != nil {
			log.Warningf(
				"Failed to parse the prefetch value \"%s\": %s",
				prefetchString,
				err.Error(),
			)
		} else {
			// Relevant code for what https://github.com/sensu/sensu is doing with this value:
			// https://github.com/sensu/sensu-transport/blob/f9c8cc0900fbef5fe9048c86116bd49efc71d801/lib/sensu/transport/rabbitmq.rb#L249
			// https://github.com/ruby-amqp/amqp/blob/9880a2b5dcfe4b27cefbdb3b3e2ea3ec58ea348a/lib/amqp/channel.rb#L998
			// https://github.com/ruby-amqp/amqp/blob/9880a2b5dcfe4b27cefbdb3b3e2ea3ec58ea348a/lib/amqp/channel.rb#L1214
			if err = t.Channel.Qos(prefetch, 0, false); err != nil {
				log.Warningf("Failed to set the prefetch value: %s", err.Error())
			}
		}
	}

	log.Noticef("RabbitMQ connection and channel opened to %s", uri)

	closeChan := make(chan *amqp.Error)
	t.Channel.NotifyClose(closeChan)

	go func() {
		<-closeChan
		t.ClosingChannel <- true
	}()

	return nil
}

func (t *RabbitMQTransport) IsConnected() bool {
	if t.Connection == nil || t.Channel == nil {
		return false
	}

	return true
}

func (t *RabbitMQTransport) Close() error {
	if t.Connection == nil {
		return errors.New("The connection is not opened")
	}

	defer func() {
		t.Channel = nil
		t.Connection = nil
	}()
	err := t.Connection.Close()

	if err != nil {
		return fmt.Errorf("Failed to close the connection: %s", err.Error())
	}

	return nil
}

func (t *RabbitMQTransport) Publish(exchangeType, exchangeName, key string, message []byte) error {
	if t.Channel == nil {
		return errors.New("The channel is not opened")
	}

	if err := t.Channel.ExchangeDeclare(exchangeName, exchangeType, false, false, false, false, nil); err != nil {
		return err
	}

	err := t.Channel.Publish(exchangeName, key, false, false, amqp.Publishing{Body: message})

	return err
}

func (t *RabbitMQTransport) Subscribe(key, exchangeName, queueName string, messageChan chan []byte, stopChan chan bool) error {
	if t.Channel == nil {
		return errors.New("The channel is not opened")
	}

	if err := t.Channel.ExchangeDeclare(
		exchangeName,
		"fanout",
		false,
		false,
		false,
		false,
		amqp.Table{},
	); err != nil {
		log.Errorf("Can't declare the exchange: %s", err.Error())
		return err
	}

	log.Infof("Exchange %s declared", exchangeName)

	if _, err := t.Channel.QueueDeclare(
		queueName,
		false,
		true,
		false,
		false,
		nil,
	); err != nil {
		log.Errorf("Can't declare the queue: %s", err.Error())
		return err
	}

	log.Infof("Queue %s declared", queueName)

	if err := t.Channel.QueueBind(queueName, key, exchangeName, false, nil); err != nil {
		log.Errorf("Can't bind the queue: %s", err.Error())
		return err
	}

	log.Noticef("Queue %s binded to %s for key %s", queueName, exchangeName, key)

	deliveryChange, err := t.Channel.Consume(queueName, "", true, false, false, false, nil)

	log.Infof("Consuming the queue %s", queueName)

	if err != nil {
		log.Errorf("Can't consume the queue: %s", err.Error())
		return err
	}

	for {
		select {
		case delivery, ok := <-deliveryChange:
			if ok {
				messageChan <- delivery.Body
			} else {
				t.ClosingChannel <- true
				return nil
			}
		case <-stopChan:
			return nil
		}
	}
}
