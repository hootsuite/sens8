package rabbitmq

import (
	"fmt"

	"github.com/upfluence/base/base_service"
	"github.com/upfluence/thrift-amqp-go/amqp_thrift"
	stdThrift "github.com/upfluence/thrift/lib/go/thrift"
	"github.com/upfluence/goutils/error_logger"
	"github.com/upfluence/goutils/thrift"
	"github.com/upfluence/goutils/thrift/handler"
)

type Endpoint struct {
	servers []*thrift.Server
}

func NewEndpoint(baseHandler *handler.Base, amqpURL string) (*Endpoint, error) {
	baseServer, err := amqp_thrift.NewTAMQPServer(
		base_service.NewBaseServiceProcessor(baseHandler),
		thrift.DefaultProtocolFactory,
		amqp_thrift.ServerOptions{
			Prefetch:     0,
			AmqpURI:      amqpURL,
			ExchangeName: handler.BASE_EXCHANGE_NAME,
			RoutingKey:   baseHandler.UnitName,
			QueueName:    fmt.Sprintf("%s-monitoring-queue", baseHandler.UnitName),
			ConsumerTag:  baseHandler.UnitName,
			Timeout:      0,
		},
	)

	if err != nil {
		return nil, err
	}

	return &Endpoint{
		servers: []*thrift.Server{thrift.NewServerFromTServer(baseServer)},
	}, nil
}

func (e *Endpoint) Mount(
	processor stdThrift.TProcessor,
	opts amqp_thrift.ServerOptions,
) error {
	s, err := amqp_thrift.NewTAMQPServer(
		processor,
		thrift.DefaultProtocolFactory,
		opts,
	)

	if err != nil {
		return err
	}

	e.servers = append(e.servers, thrift.NewServerFromTServer(s))

	return nil
}

func (e *Endpoint) Serve() error {
	errs := make(chan error)

	for _, s := range e.servers {
		go func(server *thrift.Server) {
			errs <- server.Start()
		}(s)
	}

	err := <-errs
	error_logger.DefaultErrorLogger.Capture(err, nil)

	return err
}
