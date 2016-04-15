package sd

// StaticSubscriber yields a fixed set of services.
type StaticSubscriber []Service

// Services implements Subscriber.
func (s StaticSubscriber) Services() ([]Service, error) { return s, nil }
