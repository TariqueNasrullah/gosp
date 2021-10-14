package pub_sub

// Publisher is the interface that wraps the basic Publish method.
//
// Publish method writes data to a particular topic.
type Publisher interface {
	// Publish notifies the topic with the optional data
	Publish(topic string, data interface{}) error
}

// Subscriber is the interface that wraps the basic Subscribe method.
//
// Subscribe method listens for messages of particular topic, each time
// new message is received, SubscriptionCallBackFunc method is called.
//
// Subscribe method must return a UnsubscribeFunc method that can be used
// to unsubscribe from the topic.
type Subscriber interface {
	// Subscribe to the specified topic.
	Subscribe(topic string, cb SubscriptionCallBackFunc) (UnsubscribeFunc, error)
}


// PubSub is the interface that groups the basic Publish and Subscribe methods.
type PubSub interface {
	Publisher
	Subscriber
}
