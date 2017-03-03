package sd

// Discoverer listens to a service discovery system and yields a set of
// identical instance locations. An error indicates a problem with connectivity
// to the service discovery system, or within the system itself; a subscriber
// may yield no endpoints without error.
type Discoverer interface {
	Instances() ([]string, error)
}
