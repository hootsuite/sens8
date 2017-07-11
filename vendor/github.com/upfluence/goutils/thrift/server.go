package thrift

import (
	"github.com/upfluence/thrift/lib/go/thrift"
	"github.com/upfluence/goutils/error_logger"
)

var (
	DefaultTransportFactory thrift.TTransportFactory = thrift.NewTTransportFactory()
	DefaultProtocolFactory  thrift.TProtocolFactory  = thrift.NewTBinaryProtocolFactoryDefault()
)

type Server struct {
	server thrift.TServer
}

func NewServerFromTServer(server thrift.TServer) *Server {
	return &Server{server}
}

func NewServer(processor thrift.TProcessor, transport thrift.TServerTransport) *Server {
	return &Server{
		thrift.NewTSimpleServer4(
			processor,
			transport,
			DefaultTransportFactory,
			DefaultProtocolFactory,
		),
	}
}

func (s *Server) Start() error {
	errLog := func(err error) {
		error_logger.DefaultErrorLogger.Capture(err, nil)
	}

	s.server.SetErrorLogger(errLog)

	err := s.server.Serve()
	error_logger.DefaultErrorLogger.Close()
	return err
}
