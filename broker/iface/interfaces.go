package iface

type Broker interface {
	Publish(topic string, data interface{}) error
	Consume(topic string) (<-chan []byte, error)
}
