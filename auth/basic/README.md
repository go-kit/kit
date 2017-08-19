`package auth/basic` provides a basic auth middleware [Mozilla article](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication)

## Usage

```go
import httptransport "github.com/go-kit/kit/transport/http"

httptransport.NewServer(
		endpoint.Chain(AuthMiddleware(cfg.auth.user, cfg.auth.password, "Example Realm"))(makeUppercaseEndpoint()),
		decodeMappingsRequest,
		httptransport.EncodeJSONResponse,
		httptransport.ServerBefore(httptransport.PopulateRequestContext),
	)
```

For AuthMiddleware to be able to pick up Authentication header from a http request we need to pass it through the context with something like ```httptransport.ServerBefore(httptransport.PopulateRequestContext)```