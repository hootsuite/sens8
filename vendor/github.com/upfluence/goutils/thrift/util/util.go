package util

import (
	"github.com/upfluence/thrift-amqp-go/amqp_thrift"
	"github.com/upfluence/thrift-http-go/http_thrift"
	"github.com/upfluence/thrift/lib/go/thrift"
)

var defaultProtocolFactory = thrift.NewTBinaryProtocolFactoryDefault()

func buildClient(trans thrift.TTransport, err error) (thrift.TTransport, thrift.TProtocolFactory, error) {
	if err != nil {
		return nil, nil, err
	}

	if err := trans.Open(); err != nil {
		return nil, nil, err
	}

	return trans, defaultProtocolFactory, nil
}

func BuildAMQPClient(amqpURL, exchangeName, routingKey string) (thrift.TTransport, thrift.TProtocolFactory, error) {
	return buildClient(
		amqp_thrift.NewTAMQPClient(amqpURL, exchangeName, routingKey, 0),
	)
}

func BuildHTTPClient(endpointURL string) (thrift.TTransport, thrift.TProtocolFactory, error) {
	return buildClient(http_thrift.NewTHTTPClient(endpointURL, 0, 0, 0))
}
