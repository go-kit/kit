Things that do not work:

multiple functions in an interface (goinline generates multiple NetrpcBinding structs).
Possible solutions: use different names or different pakages.

Commments, test...

No generation for
func makeEndpoint(a Add) endpoint.Endpoint.

go-generate integration (infer package from current dir)


to run:

`go run codegen/gen.go -package github.com/go-kit/kit/codegen/z -type X -w -binding=rpc,http`
