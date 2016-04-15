package sd

// Publisher publishes instance information to a service discovery system when
// an instance becomes alive and healthy, and unpublishes that information when
// the service becomes unhealthy or goes away.
//
// Publisher implementations exist for various service discovery systems. Note
// that identifying instance information (e.g. host:port) must be given via the
// concrete constructor; this interface merely signals lifecycle changes.
type Publisher interface {
	Publish()
	Unpublish()
}
