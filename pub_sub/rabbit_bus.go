package pub_sub

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"sync"
)

type rabbitBus struct {
	cfg        *RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel // don't use this channel to subscribe
	queuesMap  sync.Map
}

type subscription struct {
	subscribers     []*subscriber
	unsubscribeChan chan unsubMessage
}

type subscriber struct {
	id       uuid.UUID
	topic    string
	callback func(data interface{})
}

type unsubMessage struct {
	id    uuid.UUID
	topic string
}

func NewRabbitBus(cfg *RabbitMQConfig) (Bus, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqps://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort))
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	bus := &rabbitBus{connection: conn, channel: channel, cfg: cfg}

	if err := bus.exchangeDeclare(cfg.AMQP.Exchange, cfg.AMQP.ExchangeType); err != nil {
		return nil, err
	}
	return bus, nil
}

func (rb *rabbitBus) exchangeDeclare(name, kind string) error {
	return rb.channel.ExchangeDeclare(
		name,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (rb *rabbitBus) Notify(topic string, data interface{}) error {
	bt, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rb.channel.Publish(
		rb.cfg.AMQP.Exchange,
		topic,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        bt,
		},
	)
}

func (rb *rabbitBus) Subscribe(topic string, cb func(data interface{})) (func(), error) {
	id := uuid.New()

	if _, ok := rb.queuesMap.Load(topic); !ok {
		newSubscribersSlice := make([]*subscriber, 0)
		newSubscriber := subscriber{
			id:       id,
			topic:    topic,
			callback: cb,
		}
		newSubscribersSlice = append(newSubscribersSlice, &newSubscriber)

		unsubChan := make(chan unsubMessage)
		newSubscription := subscription{
			subscribers:     newSubscribersSlice,
			unsubscribeChan: unsubChan,
		}
		rb.queuesMap.Store(topic, &newSubscription)

		q, err := rb.channel.QueueDeclare(
			"",
			false,
			true,
			true,
			false,
			nil,
		)
		if err != nil {
			return nil, err
		}

		err = rb.channel.QueueBind(
			q.Name,
			topic,
			rb.cfg.AMQP.Exchange,
			false,
			nil,
		)
		if err != nil {
			return nil, err
		}

		msgs, err := rb.channel.Consume(
			q.Name,
			"",
			true,
			false,
			false,
			false,
			nil)
		if err != nil {
			return nil, err
		}

		go func() {
			for {
				select {
				case unsubMsg := <-unsubChan:
					if val, ok := rb.queuesMap.Load(unsubMsg.topic); ok {
						subs := val.(*subscription)
						subscribers := subs.subscribers

						for i, sub := range subscribers {
							if sub.id == unsubMsg.id {
								subscribers = append(subscribers[:i], subscribers[i+1:]...)
								subs.subscribers = subscribers
								rb.queuesMap.Store(topic, subs)
								break
							}
						}
					}
				case data := <-msgs:
					if data.Body != nil {
						if val, ok := rb.queuesMap.Load(topic); ok {
							subs := val.(*subscription)
							if subs.subscribers != nil {
								for _, subscriber := range subs.subscribers {
									subscriber.callback(string(data.Body))
								}
							}
						}
					}
				}
			}
		}()
	} else {
		if val, ok := rb.queuesMap.Load(topic); ok {
			subs := val.(*subscription)
			subscribers := subs.subscribers
			newSubscriber := subscriber{
				id:       id,
				topic:    topic,
				callback: cb,
			}
			subscribers = append(subscribers, &newSubscriber)
			subs.subscribers = subscribers
			rb.queuesMap.Store(topic, subs)
		}
	}

	return func() {
		msg := unsubMessage{
			id:    id,
			topic: topic,
		}
		if val, ok := rb.queuesMap.Load(topic); ok {
			queue := val.(*subscription)
			queue.unsubscribeChan <- msg
		}
	}, nil
}

func (rb *rabbitBus) DeleteQueue(id string) error {
	channel, err := rb.connection.Channel()
	if err != nil {
		return err
	}

	_, err = channel.QueueDelete(id, false, false, true)
	if err != nil {
		return err
	}
	return nil
}
