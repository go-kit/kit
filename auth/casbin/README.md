# package auth/casbin

`package auth/casbin` provides endpoint middleware for authorization with
[Casbin](https://casbin.org/).

## Usage

`NewEnforcer(subject, object, action)` returns an `endpoint.Middleware`. The
middleware reads the Casbin model and policy from the request context:

- `CasbinModelContextKey`: a path to a model file, or a `casbin/model` `Model`
- `CasbinPolicyContextKey`: a path to a policy file, or a `casbin/persist` `Adapter`

On a successful check, the generated Enforcer is stored in the context under
`CasbinEnforcerContextKey`. If the subject is not allowed to perform the
action on the object, the middleware returns `ErrUnauthorized`.

```go
import (
	"github.com/go-kit/kit/auth/casbin"
	"github.com/go-kit/kit/endpoint"
)

func main() {
	var exampleEndpoint endpoint.Endpoint
	{
		exampleEndpoint = MakeExampleEndpoint(service)
		exampleEndpoint = casbin.NewEnforcer("alice", "data1", "read")(exampleEndpoint)
	}
}
```

The model and policy must already be on the context when the middleware runs.
For HTTP servers, add them with `ServerBefore`. `PopulateRequestContext` can
be used alongside that to copy standard request values into the context.

```go
import (
	"context"
	"net/http"

	"github.com/go-kit/kit/auth/casbin"
	httptransport "github.com/go-kit/kit/transport/http"
)

func main() {
	httptransport.NewServer(
		casbin.NewEnforcer("alice", "data1", "read")(makeExampleEndpoint()),
		decodeRequest,
		httptransport.EncodeJSONResponse,
		httptransport.ServerBefore(
			httptransport.PopulateRequestContext,
			func(ctx context.Context, _ *http.Request) context.Context {
				ctx = context.WithValue(ctx, casbin.CasbinModelContextKey, "testdata/basic_model.conf")
				ctx = context.WithValue(ctx, casbin.CasbinPolicyContextKey, "testdata/basic_policy.csv")
				return ctx
			},
		),
	)
}
```
