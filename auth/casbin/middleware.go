package casbin

import (
	"context"
	"errors"

	stdcasbin "github.com/casbin/casbin/v2"
	"github.com/openmesh/kit/endpoint"
)

type contextKey string

const (
	// CasbinModelContextKey holds the key to store the access control model
	// in context, it can be a path to configuration file or a casbin/model
	// Model.
	CasbinModelContextKey contextKey = "CasbinModel"

	// CasbinPolicyContextKey holds the key to store the access control policy
	// in context, it can be a path to policy file or an implementation of
	// casbin/persist Adapter interface.
	CasbinPolicyContextKey contextKey = "CasbinPolicy"

	// CasbinEnforcerContextKey holds the key to retrieve the active casbin
	// Enforcer.
	CasbinEnforcerContextKey contextKey = "CasbinEnforcer"
)

var (
	// ErrModelContextMissing denotes a casbin model was not passed into
	// the parsing of middleware's context.
	ErrModelContextMissing = errors.New("CasbinModel is required in context")

	// ErrPolicyContextMissing denotes a casbin policy was not passed into
	// the parsing of middleware's context.
	ErrPolicyContextMissing = errors.New("CasbinPolicy is required in context")

	// ErrUnauthorized denotes the subject is not authorized to do the action
	// intended on the given object, based on the context model and policy.
	ErrUnauthorized = errors.New("Unauthorized Access")
)

// NewEnforcer checks whether the subject is authorized to do the specified
// action on the given object. If a valid access control model and policy
// is given, then the generated casbin Enforcer is stored in the context
// with CasbinEnforcer as the key.
func NewEnforcer[Request, Response any](
	subject string, object interface{}, action string,
) endpoint.Middleware[Request, Response] {
	return func(next endpoint.Endpoint[Request, Response]) endpoint.Endpoint[Request, Response] {
		return func(ctx context.Context, request Request) (response Response, err error) {
			casbinModel := ctx.Value(CasbinModelContextKey)
			casbinPolicy := ctx.Value(CasbinPolicyContextKey)
			enforcer, err := stdcasbin.NewEnforcer(casbinModel, casbinPolicy)
			if err != nil {
				return *new(Response), err
			}

			ctx = context.WithValue(ctx, CasbinEnforcerContextKey, enforcer)
			ok, err := enforcer.Enforce(subject, object, action)
			if err != nil {
				return *new(Response), err
			}
			if !ok {
				return *new(Response), ErrUnauthorized
			}

			return next(ctx, request)
		}
	}
}
