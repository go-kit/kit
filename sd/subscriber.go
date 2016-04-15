package sd

// Subscriber listens to a service discovery system and yields a set of
// identical services on demand. Typically, this means a set of identical
// instances of a microservice. An error indicates a problem with connectivity
// to the service discovery system, or within the system itself; a subscriber
// may yield no services without error.
type Subscriber interface {
	Services() ([]Service, error)
}
