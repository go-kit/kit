package sd

// FixedDiscoverer yields a fixed set of instances.
type FixedDiscoverer []string

// Instances implements Discoverer.
func (d FixedDiscoverer) Instances() ([]string, error) { return d, nil }
