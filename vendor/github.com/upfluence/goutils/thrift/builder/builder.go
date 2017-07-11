package builder

import (
	"errors"
	"time"

	"github.com/streadway/amqp"
	"github.com/upfluence/base/service/thrift/transport/http"
	"github.com/upfluence/base/service/thrift/transport/rabbitmq"
	"github.com/upfluence/base/service/thrift_service"
	"github.com/upfluence/thrift-amqp-go/amqp_thrift"
	"github.com/upfluence/thrift/lib/go/thrift"
)

const (
	defaultOpenTimeout = 30 * time.Second
	defaultRabbitMQURL = "amqp://localhost:5672/%2f"
)

var (
	errNoTransport        = errors.New("No transport")
	errHTTPTransportNoURL = errors.New("No URL provided")
	protocolFactory       = thrift.NewTBinaryProtocolFactoryDefault()
)

type Builder interface {
	Build(service *thrift_service.Service) (thrift.TTransport, thrift.TProtocolFactory, error)
}

type builder struct {
	opts Options
}

type RabbitMQOptions struct {
	Connection                  *amqp.Connection
	Channel                     *amqp.Channel
	URL, QueueName, ConsumerTag string
	OnewayMode                  bool
}

type Options struct {
	RabbitMQOptions        *RabbitMQOptions
	DefaultProtocolFactory thrift.TProtocolFactory
}

func NewBuilder(opts *Options) Builder {
	var o Options

	if opts != nil {
		o = *opts
	}

	return &builder{opts: o}
}

func (b *builder) Build(service *thrift_service.Service) (thrift.TTransport, thrift.TProtocolFactory, error) {
	var (
		transport thrift.TTransport
		err       error

		protFactory thrift.TProtocolFactory = protocolFactory
	)

	if fact := b.opts.DefaultProtocolFactory; fact != nil {
		protFactory = fact
	}

	if trans := service.Transport; trans != nil {
		if t := trans.HttpTransport; t != nil {
			transport, err = b.buildHTTPTransport(t)
		} else if t := trans.RabbitmqTransport; t != nil {
			transport, err = b.buildRabbitMQTransport(t)
		}
	}

	if transport == nil && err == nil {
		err = errNoTransport
	}

	return transport, protFactory, err
}

func (b *builder) buildHTTPTransport(transport *http.Transport) (thrift.TTransport, error) {
	if url := transport.GetUrl(); url == "" {
		return nil, errHTTPTransportNoURL
	}

	return thrift.NewTHttpPostClient(transport.GetUrl())
}

func (b *builder) buildRabbitMQTransport(transport *rabbitmq.Transport) (thrift.TTransport, error) {
	var rabbitMQURL = defaultRabbitMQURL

	if t := b.opts.RabbitMQOptions; t != nil {
		if t.URL != "" {
			rabbitMQURL = t.URL
		}

		return amqp_thrift.NewTAMQPClientFromConnAndQueue(
			t.Connection,
			t.Channel,
			transport.GetExchangeName(),
			transport.GetRoutingKey(),
			t.ConsumerTag,
			t.QueueName,
			defaultOpenTimeout,
			t.OnewayMode,
		)
	}

	return amqp_thrift.NewTAMQPClient(
		rabbitMQURL,
		transport.GetExchangeName(),
		transport.GetRoutingKey(),
		defaultOpenTimeout,
	)
}
