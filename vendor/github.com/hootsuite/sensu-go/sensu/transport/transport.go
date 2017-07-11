package transport

type Transport interface {
	Connect() error
	IsConnected() bool
	Close() error
	Publish(exchangeType, exchangeName, key string, message []byte) error
	Subscribe(key, exchangeName, queueName string, messageChan chan []byte, stopChan chan bool) error
	GetClosingChan() chan bool
}
