package sd

import "github.com/go-kit/kit/service"

// FixedSubscriber yields a fixed set of services.
type FixedSubscriber []service.Service

// Services implements Subscriber.
func (s FixedSubscriber) Services() ([]service.Service, error) { return s, nil }
