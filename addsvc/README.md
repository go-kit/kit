# addsvc

addsvc is an example service, used to illustrate the mechanics of Go kit. It
exposes a single method to add two integers on a variety of transports and
endpoints.

## Highlights

### Configuration via flags

Go kit has no strong opinions about how to pass configuration to your service.
If your organization has established conventions to pass configuration into
your service, Go kit won't stand in your way. That said, package flag is a
good default: it's simple, well-understood, and provides a self-documenting
configuration surface area. Keeping with
 [best practices](http://peter.bourgon.org/go-in-production/#configuration), flags
are defined in func main.

### Declarative composition

Go kit strongly favors explicit, declarative composition of interacting
components via a comprehensive func main. Time spent in keystrokes is made up
many, many times over when returning to the code and understanding exactly
what's happening, without having to unravel indirections or abstractions.

### Multiple transports

Go kit treats transports — HTTP, Thrift, gRPC, etc. — as pluggable. The same
service can be exposed on any, or multiple, available transports. The addsvc
example demonstrates how to make the same business logic available over
multiple transports simultaneously.

### Daemonizing

Go kit has no strong opinions about how to daemonize, supervise, or run your
service. If your organization has established conventions for running
services. Go kit won't stand in your way. Go kit services run equally well as
manually-copied binaries; applications provisioned with configuration
management tools like [Chef][], [Puppet][], or [Ansible][]; in containers like
[Docker][] or [rkt][]; or as part of a comprehensive scheduling platform like
[Kubernetes][], [Mesos][], [OpenStack][], [Deis][], etc.

[Chef]: https://www.chef.io
[Puppet]: https://puppetlabs.com
[Ansible]: http://www.ansible.com
[Docker]: http://docker.com
[rkt]: https://github.com/coreos/rkt
[Kubernetes]: http://kubernetes.io
[Mesos]: https://mesosphere.com
[OpenStack]: https://www.openstack.org
[Deis]: http://deis.io

## Server

To build and run addsvc,

```
go install
addsvc
```

## Client

addsvc comes with an example client, [addcli][].

[addcli]: https://github.com/go-kit/kit/blob/master/addsvc/client/addcli/main.go

```
$ cd client/addcli
$ go install
$ addcli
```

