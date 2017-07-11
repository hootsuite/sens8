package serializer

import "github.com/upfluence/thrift/lib/go/thrift"

type TStruct interface {
	Write(thrift.TProtocol) error
	Read(thrift.TProtocol) error
}
