package pub_sub

type (
	SubscriptionCallBackFunc func(data interface{})
	UnsubscribeFunc func() error
)