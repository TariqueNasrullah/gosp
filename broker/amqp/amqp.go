package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/TariqueNasrullah/gosp/broker/iface"
	"github.com/TariqueNasrullah/gosp/config"
	"github.com/streadway/amqp"
)

type Broker struct {
	cfg        config.AMQPConfig
	connection amqp.Connection
}

func New() iface.Broker {
	return &Broker{}
}

func (b *Broker) Publish(topic string, data interface{}) error {
	bt, err := json.Marshal(data)
	if err != nil {
		return err
	}

	channel, err := b.connection.Channel()
	if err != nil {
		return err
	}

	// Enable publish confirmations
	if err = channel.Confirm(false); err != nil {
		return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
	}

	channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	b.connection.NotifyClose(make(chan *amqp.Error, 1))

	return channel.Publish(
		b.cfg.Exchange,
		topic,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bt,
		},
	)
}

func (b *Broker) Consume(topic string) (<-chan []byte, error) {
	panic("implement me")
}
