package rabbitmq

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

type mockAMQPChannel struct {
	qos struct {
		prefetchCount, prefetchSize int
		global                      bool
	}
}

func (*mockAMQPChannel) Consume(
	string,
	string,
	bool,
	bool,
	bool,
	bool,
	amqp.Table,
) (<-chan amqp.Delivery, error) {
	return nil, nil
}

func (*mockAMQPChannel) ExchangeDeclare(
	string,
	string,
	bool,
	bool,
	bool,
	bool,
	amqp.Table,
) error {
	return nil
}

func (*mockAMQPChannel) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	// We need to close the channel here, in order to prevent the goroutine at
	// the end of transport.Connect() from blocking indefinitely.
	close(c)
	return nil
}

func (*mockAMQPChannel) Publish(string, string, bool, bool, amqp.Publishing) error {
	return nil
}

func (c *mockAMQPChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	c.qos.prefetchCount = prefetchCount
	c.qos.prefetchSize = prefetchSize
	c.qos.global = global

	return nil
}

func (*mockAMQPChannel) QueueBind(string, string, string, bool, amqp.Table) error {
	return nil
}

func (*mockAMQPChannel) QueueDeclare(
	string,
	bool,
	bool,
	bool,
	bool,
	amqp.Table,
) (amqp.Queue, error) {
	return amqp.Queue{}, nil
}

type mockAMQPConnection struct {
	heartbeat time.Duration
	channel   *mockAMQPChannel
}

func (c *mockAMQPConnection) Channel() (AMQPChannel, error) {
	return c.channel, nil
}

func (*mockAMQPConnection) Close() error {
	return nil
}

func mockAMQPDialer(url string) (AMQPConnection, error) {
	return mockAMQPDialerConfig(url, amqp.Config{})
}

func mockAMQPDialerConfig(url string, config amqp.Config) (AMQPConnection, error) {
	return &mockAMQPConnection{channel: &mockAMQPChannel{}, heartbeat: config.Heartbeat}, nil
}

func getDummyTransportConfig(heartbeat int, prefetch int) *TransportConfig {
	config, _ := NewTransportConfig("amqp://guest:guest@localhost:5672/%2F")

	if heartbeat == 0 {
		return config
	}

	config.Heartbeat = json.Number(strconv.Itoa(heartbeat))

	if prefetch == 0 {
		return config
	}

	config.Prefetch = json.Number(strconv.Itoa(prefetch))

	return config
}

var transportConnectTestScenarios = []struct {
	config            *TransportConfig
	expectedHeartbeat time.Duration
	expectedPrefetch  int
}{
	{
		getDummyTransportConfig(0, 0),
		10 * time.Second,
		0,
	},
	{
		getDummyTransportConfig(41, 42),
		time.Duration(41 * time.Second),
		42,
	},
}

func TestTransportConnect(t *testing.T) {
	for _, scenario := range transportConnectTestScenarios {
		transport := &RabbitMQTransport{
			ClosingChannel: make(chan bool),
			Configs:        []*TransportConfig{scenario.config},
			dialer:         mockAMQPDialer,
			dialerConfig:   mockAMQPDialerConfig,
		}

		err := transport.Connect()

		validateError(err, nil, t)

		connection := transport.Connection.(*mockAMQPConnection)

		if connection.heartbeat != scenario.expectedHeartbeat {
			t.Errorf("Expected heartbeat to be \"%s\" but got \"%s\" instead",
				scenario.expectedHeartbeat,
				connection.heartbeat,
			)
		}

		if transport.Channel != connection.channel {
			t.Errorf(
				"Expected transport.Channel to be \"%+v\" but got \"%+v\" instead",
				connection.channel,
				transport.Channel,
			)
		}

		if connection.channel.qos.prefetchCount != scenario.expectedPrefetch {
			t.Errorf(
				"Expected prefetch value to be \"%d\" but got \"%d\" instead",
				scenario.expectedPrefetch,
				connection.channel.qos.prefetchCount,
			)
		}

		if connection.channel.qos.prefetchSize != 0 {
			t.Errorf(
				"Expected prefetch size to be \"%d\" but got \"%d\" instead",
				0,
				connection.channel.qos.prefetchSize,
			)
		}

		if connection.channel.qos.global {
			t.Errorf(
				"Expected qos global to be \"%t\" but got \"%t\" instead",
				false,
				connection.channel.qos.global,
			)
		}

		channelClosed := <-transport.ClosingChannel

		if !channelClosed {
			t.Error("Failed to close channel")
		}

		err = transport.Close()

		if err != nil {
			t.Fatalf("Expected error to be nil but got \"%s\" instead", err)
		}

		if transport.Channel != nil {
			t.Errorf(
				"Expected channel to be nil but got \"%+v\" instead",
				transport.Channel,
			)
		}

		if transport.Connection != nil {
			t.Errorf(
				"Expected connection to be nil but got \"%+v\" instead",
				transport.Connection,
			)
		}
	}
}

var errFailedToConnect = errors.New("Dummy dialer error")

func mockAMQPDialerError(url string) (AMQPConnection, error) {
	return nil, errFailedToConnect
}

func mockAMQPDialerConfigError(url string, conf amqp.Config) (AMQPConnection, error) {
	return nil, errFailedToConnect
}

func TestTransportConnectError(t *testing.T) {
	transport := &RabbitMQTransport{
		ClosingChannel: make(chan bool),
		Configs:        []*TransportConfig{getDummyTransportConfig(0, 0)},
		dialer:         mockAMQPDialerError,
		dialerConfig:   mockAMQPDialerConfigError,
	}

	err := transport.Connect()

	validateError(err, errFailedToConnect, t)
}

func TestTransportSubscribe(t *testing.T) {
	transport := &RabbitMQTransport{
		ClosingChannel: make(chan bool),
		Configs:        []*TransportConfig{getDummyTransportConfig(0, 0)},
		dialer:         mockAMQPDialer,
		dialerConfig:   mockAMQPDialerConfig,
	}

	err := transport.Connect()

	validateError(err, nil, t)

	stopChan := make(chan bool)

	waitForSubscribe := make(chan bool)

	go func() {
		err = transport.Subscribe("", "", "", nil, stopChan)

		validateError(err, nil, t)

		waitForSubscribe <- true
	}()

	stopChan <- true

	<-waitForSubscribe
}
