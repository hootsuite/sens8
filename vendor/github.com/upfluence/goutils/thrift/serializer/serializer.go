package serializer

import "github.com/upfluence/thrift/lib/go/thrift"

type TSerializer struct {
	transport *thrift.TMemoryBuffer
	protocol  thrift.TProtocol
}

func NewTSerializer(protocolFactory thrift.TProtocolFactory) *TSerializer {
	var transport = thrift.NewTMemoryBufferLen(1024)

	protocol := protocolFactory.GetProtocol(transport)

	return &TSerializer{
		transport,
		protocol,
	}
}

func (t *TSerializer) ReadString(msg TStruct, s string) error {
	return t.Read(msg, []byte(s))
}

func (t *TSerializer) Read(msg TStruct, b []byte) error {
	if _, err := t.transport.Write(b); err != nil {
		return err
	}

	if err := msg.Read(t.protocol); err != nil {
		return err
	}

	return nil
}

func (t *TSerializer) writeAndFlush(msg TStruct) error {
	t.transport.Reset()

	if err := msg.Write(t.protocol); err != nil {
		return err
	}

	if err := t.protocol.Flush(); err != nil {
		return err
	}

	if err := t.transport.Flush(); err != nil {
		return err
	}

	return nil
}

func (t *TSerializer) Write(msg TStruct) ([]byte, error) {
	if err := t.writeAndFlush(msg); err != nil {
		return []byte{}, err
	}

	return t.transport.Bytes(), nil
}

func (t *TSerializer) WriteString(msg TStruct) (string, error) {
	if err := t.writeAndFlush(msg); err != nil {
		return "", err
	}

	return t.transport.String(), nil
}
