# package auth/jwt

`package auth/jwt` provides a set of interfaces for service authorization through [JSON Web Tokens](https://jwt.io/).

## Usage

NewParser takes a key function and an expected signing method and returns an `endpoint.Middleware`. 
The middleware will parse a token passed into the context via the `jwt.JWTTokenContextKey`. 
If the token is valid, any claims will be added to the context via the `jwt.JWTClaimsContextKey`.

```go
import (
	stdjwt "github.com/dgrijalva/jwt-go"
    
    "github.com/go-kit/kit/auth/jwt"
    "github.com/go-kit/kit/endpoint"
)

func main() {
	var exampleEndpoint endpoint.Endpoint
	{
		keyFunc := func(token *stdjwt.Token) (interface{}, error) { return []byte("SigningString"), nil }
		jwtParser := jwt.NewParser(keyFunc, stdjwt.SigningMethodHS256)
        
		exampleEndpoint = MakeExampleEndpoint(service)
		exampleEndpoint = jwtParser(exampleEndpoint)
	}
}
```

NewSigner takes a JWT key id header, the signing key, signing method, and a claims object. It returns an `endpoint.Middleware`.
The middleware will build the token string and add it to the context via the `jwt.JWTTokenContextKey`.

```go
import (
	stdjwt "github.com/dgrijalva/jwt-go"
    
    "github.com/go-kit/kit/auth/jwt"
    "github.com/go-kit/kit/endpoint"
)

func main() {
	var exampleEndpoint endpoint.Endpoint
	{
		jwtSigner := jwt.NewSigner("kid-header", []byte("SigningString"), stdjwt.SigningMethodHS256, jwt.Claims{})
        
		exampleEndpoint = grpctransport.NewClient(
        	. // build client endpoint here
			.
			.
		).Endpoint()

		exampleEndpoint = jwtSigner(exampleEndpoint)
	}
}
```

In order for the parser and the signer to work, the authorization headers need to be passed between the request and the context.
`ToHTTPContext()`, `FromHTTPContext()`, `ToGRPCContext()`, and `FromGRPCContext()` are given as helpers to do this.
These function impliment the correlating transport's RequestFunc interface and can be passes as ClientBefore or ServerBefore options.

Example of use in a client:

```go
import (
	stdjwt "github.com/dgrijalva/jwt-go"
    
    "github.com/go-kit/kit/auth/jwt"
    "github.com/go-kit/kit/endpoint"
)

func main() {

    options := []httptransport.ClientOption{}
	var exampleEndpoint endpoint.Endpoint
	{
		jwtSigner := jwt.NewSigner("kid-header", []byte("SigningString"), stdjwt.SigningMethodHS256, jwt.Claims{})
       
		options = append(options, httptransport.ClientBefore(jwt.FromGRPCContext()))
		exampleEndpoint = grpctransport.NewClient(
        	. // build client endpoint here
			.
			options....
		).Endpoint()

		exampleEndpoint = jwtSigner(exampleEndpoint)
	}
}
```

Example of use in a server:

```go
import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

func MakeGRPCServer(ctx context.Context, endpoints Endpoints, logger log.Logger) pb.ExampleServer {
	options := []grpctransport.ServerOption{grpctransport.ServerErrorLogger(logger)}

	return &grpcServer{
		createUser: grpctransport.NewServer(
			ctx,
			endpoints.CreateUserEndpoint,
			DecodeGRPCCreateUserRequest,
			EncodeGRPCCreateUserResponse,
			append(options, grpctransport.ServerBefore(jwt.ToGRPCContext()))...,
		),
		getUser: grpctransport.NewServer(
			ctx,
			endpoints.GetUserEndpoint,
			DecodeGRPCGetUserRequest,
			EncodeGRPCGetUserResponse,
			options...,
		),
	}
}
```
