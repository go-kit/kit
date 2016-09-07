package lb

import (
    "fmt"
    "strings"
    "time"

    "golang.org/x/net/context"

    "github.com/go-kit/kit/endpoint"
)

// Callback function that indicates the current attempt count and the error encountered.
//  Should return whether the Retry function should continue trying, and a custom 
//  error message if desired. The error message may be nil, but a true/false 
//  is always expected. In all cases if the error message is supplied, the 
//  current error will be replaced.
type callback func(int, string) (bool, *string)

// Retry wraps a service load balancer and returns an endpoint oriented load
// balancer for the specified service method.
// Requests to the endpoint will be automatically load balanced via the load
// balancer. Requests that return errors will be retried until they succeed,
// up to max times, or until the timeout is elapsed, whichever comes first.
func Retry(max int, timeout time.Duration, b Balancer) endpoint.Endpoint {
    if b == nil {
        panic("nil Balancer")
    }
    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        var (
            newctx, cancel = context.WithTimeout(ctx, timeout)
            responses        = make(chan interface{}, 1)
            errs                  = make(chan error, 1)
            a                      = []string{}
        )
        defer cancel()
        for i := 1; i <= max; i++ {
            go func() {
                e, err := b.Endpoint()
                if err != nil {
                    errs <- err
                    return
                }
                response, err := e(newctx, request)
                if err != nil {
                    errs <- err
                    return
                }
                responses <- response
            }()

            select {
            case <-newctx.Done():
                return nil, newctx.Err()
            case response := <-responses:
                return response, nil
            case err := <-errs:
                a = append(a, err.Error())
                continue
            }
        }
        return nil, fmt.Errorf("retry attempts exceeded (%s)", strings.Join(a, "; "))
    }
}

func RetryWithCallback(max int, timeout time.Duration, b Balancer, cb callback) endpoint.Endpoint {
    if cb == nil {
        panic("nil callback")
    }
    if b == nil {
        panic("nil Balancer")
    }
    return func(ctx context.Context, request interface{}) (response interface{}, err error) {
        var (
            newctx, cancel = context.WithTimeout(ctx, timeout)
            responses        = make(chan interface{}, 1)
            errs                  = make(chan error, 1)
            a                      = []string{}
        )
        defer cancel()
        for i := 1; i <= max; i++ {
            go func() {
                e, err := b.Endpoint()
                if err != nil {
                    errs <- err
                    return
                }
                response, err := e(newctx, request)
                if err != nil {
                    errs <- err
                    return
                }
                responses <- response
            }()

            select {
            case <-newctx.Done():
                return nil, newctx.Err()
            case response := <-responses:
                return response, nil
            case err := <-errs:
                cont, cbErr := cb(i, err.Error())
                if !cont {
                    if cbErr == nil {
                        return nil, fmt.Errorf("retry attempts exceeded (%s)", strings.Join(a, "; "))   
                    } else {
                        return nil, fmt.Errorf(*cbErr) 
                    }
                }
                currentErr := err.Error()
                if cbErr != nil {
                    currentErr = *cbErr
                } 
                a = append(a, currentErr)
                continue
            }
        }
        return nil, fmt.Errorf("retry attempts exceeded (%s)", strings.Join(a, "; "))
    }
}
