package amqputil

import (
	"time"

	"github.com/streadway/amqp"
	"github.com/upfluence/goutils/log"
)

const defaultBackoff = 5 * time.Second

func BuildConnection(
	uri string,
) (*amqp.Connection, *amqp.Channel, chan<- bool) {
	var closeChan = make(chan bool)

	connection, channel := buildConnection(uri)

	go func() {
		var amqpChan = make(chan *amqp.Error)

		for {
			channel.NotifyClose(amqpChan)

			select {
			case <-closeChan:
				log.Warningf("amqputil: watcher closing...")
				return
			case <-amqpChan:
				log.Errorf("amqputil: Channel closed")

				connection, channel = buildConnection(uri)
			}
		}
	}()

	return connection, channel, closeChan
}

func buildConnection(uri string) (*amqp.Connection, *amqp.Channel) {
	log.Warningf("Connecting to %s...", uri)

	connection, err := amqp.Dial(uri)

	if err != nil {
		log.Errorf("amqputil: dial error: %s", err.Error())
		time.Sleep(defaultBackoff)

		return buildConnection(uri)
	}

	channel, err := connection.Channel()

	if err != nil {
		log.Errorf("amqputil: channel error: %s", err.Error())
		time.Sleep(defaultBackoff)

		return buildConnection(uri)
	}

	log.Warningf("Connection to %s opened", uri)
	return connection, channel
}
