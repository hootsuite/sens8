package serializer

import "github.com/upfluence/thrift/lib/go/thrift"

var defaultFactory = thrift.NewTJSONProtocolFactory()

type TSerializerFactory struct {
	protocolFactory thrift.TProtocolFactory
}

func NewTSerializerFactory(protocolFactory thrift.TProtocolFactory) *TSerializerFactory {
	return &TSerializerFactory{protocolFactory}
}

func NewDefaultTSerializerFactory() *TSerializerFactory {
	return &TSerializerFactory{defaultFactory}
}

func (f *TSerializerFactory) GetSerializer() *TSerializer {
	return NewTSerializer(f.protocolFactory)
}
