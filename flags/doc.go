// Package flags provides utilities related to feature flags.
//
// Flags/Toggles are dependencies, and should be passed to the components that
// need them in the same way you'd construct and pass a database handle, or
// reference to another component. Instantiate flags in your func main, using
// whichever concrete implementation is appropriate for your organization.
package flags
