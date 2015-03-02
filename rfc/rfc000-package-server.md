---
RFC: 000
Author: Matt Heath <matt@mattheath.com>
Status: Accepted
---

# Package Server

## Motivation

[http://peter.bourgon.org/go-kit/#package-server](http://peter.bourgon.org/go-kit/#package-server)

Package server is probably the biggest and most important component of the toolkit. Ideally, we should be able to write our services as implementations of normal, nominal Go interfaces, and delegate integration with the environment to the server package. The package should encode and enforce conventions for server-side concerns, like health checks, system-wide request tracing, connection management, backpressure and throttling, and so on. For each of those topics, it should provide interfaces for different, pluggable strategies. It should integrate with service discovery, and work equally well over multiple transports. Considerable prior art exists in the form of Finagle, Karyon (Netflix's application service library), and likely many more.

## Scope

Sections with a ★ are considered particularly volatile, and may change significantly in the future.

### Endpoints

*   An endpoint is defined as a handler interface which receives a request and returns a response or an error to the client.
*   A server SHALL have one or more endpoints.
*   An endpoint SHALL accept a request context and propagate this - allowing access to request-specific values such as the identity of the end user, authorization tokens, and the request's deadline.
*   An endpoint SHALL respect request cancellation if or when the request's deadline expires.
*   A server MAY expose information about its endpoints, to allow integration with additional tools.

### Contexts

*   Requests SHALL be executed within a request [context](https://blog.golang.org/context), which the server will pass through the request chain.

### Throttling & Backpressure

*   A server MAY throttle inbound requests and reject requests from clients based on a number of factors.
*   A server MAY respond with either an Out Of Capacity error, or a Rate Limit Exceeded error when rejecting requests.
*   A server MAY limit the total number of concurrent requests it can serve.
*   A server MAY impose rate limits on specific clients.
*   Rate limit behaviour MAY range from minimum request intervals, to time based, or leaky bucket algorithms.
*   A server MAY implement a pluggable throttle interface, allowing richer implementations - such as an implementation which shares information across instances of the service.

### SLAs & SLIs

*   A server MAY report its contractual SLA per endpoint to a discovery system, allowing clients to estimate response time.
*	A server MAY expose its actual SLI per endpoint, allowing third-parties to reason about healthiness.

### Healthchecks ★

*   A server SHALL accept registration of healthchecks with a defined interface.
*   A server MAY register default healthchecks to report the health of built in components of the server.
*   A server SHALL register an endpoint which can be queried to obtain healthcheck information and status.
*   A server MAY publish these healthcheck statuses via a pluggable transport.

### Service Discovery

*   A server SHALL register itself with a service discovery mechanism on startup.
*   A server SHALL attempt to deregister itself with a service discovery mechanism on shutdown.
*   The discovery mechanism SHALL be interchangeable, and satisfy a defined interface, however the mechanism itself is beyond the scope of this RFC.

### Request Tracing

*   Requests received by the server which are accompanied with tracing information SHALL respect this information and pass this information onto other sub-requests initiated by clients within the server.
*   The request tracing mechanism SHALL be interchangeable, and satisfy a defined interface, however the mechanism is beyond the scope of this RFC.

### Transport

*   A server SHALL receive and respond to requests via a Transport.
*   The Transport mechanism SHALL be interchangeable, and satisfy a defined interface, however the mechanism of the transport is beyond the scope of this RFC.

### Codec ★

*   A server SHALL encode and decode requests and responses via an interchangeable Codec.
*   A server MAY support multiple encodings, and use the appropriate Codec as indicated by the transport.
*   A server MAY indicate to the transport the encoding used, allowing clients to easily decode the response.

## Implementation

To be defined.

## Further Reading

*	[Your Server as a Function](http://monkey.org/~marius/funsrv.pdf) - Marius Eriksen
*	[Finagle](https://twitter.github.io/finagle/) - Twitter
*	[Karyon](https://github.com/Netflix/karyon) - Netflix
*	[State of the Art in Microservices](https://www.slideshare.net/adriancockcroft/dockercon-state-of-the-art-in-microservices) - Adrian Cockcroft
