package pub_sub

// Bus is an interface used for pub-sub like interactions
type Bus interface {
	// Subscribe to the specified topic.
	Subscribe(topic string, cb func(data interface{})) (func(), error)

	// Notify notifies the topic with the optional data
	Notify(topic string, data interface{}) error

	// DeleteQueue deletes a queue
	DeleteQueue(id string) error
}
