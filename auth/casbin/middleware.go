package casbin

import (
	"context"
	"errors"

	stdcasbin "github.com/casbin/casbin"
	"github.com/go-kit/kit/endpoint"
)

var (
	// ErrModelContextMissing required CasbinModel
	ErrModelContextMissing = errors.New("CasbinModel is required in context")
	// ErrPolicyContextMissing required CasbinPolicy
	ErrPolicyContextMissing = errors.New("CasbinPolicy is required in context")
	// ErrUnauthorized describes unauthorized access
	ErrUnauthorized = errors.New("Unauthorized Access")
)

type contextKey string

const (
	// CasbinModelContextKey key to store the model, can be a file or casbin model
	// a model file e.g. "path/to/basic_model.conf"
	CasbinModelContextKey contextKey = "CasbinModel"
	// CasbinPolicyContextKey key to store the policy, can be a file or casbin policy adapter
	// a policy file e.g. "path/to/basic_policy.csv"
	CasbinPolicyContextKey contextKey = "CasbinPolicy"
	// CasbinEnforcerContextKey key where the active enforcer can be retrieved
	CasbinEnforcerContextKey contextKey = "CasbinEnforcer"
)

// NewEnforcer installs casbin enforcer into the context
// while also checking the authorization for the corresponding subject, object, and action
func NewEnforcer(subject string, object interface{}, action string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			casbinModel := ctx.Value(CasbinModelContextKey)
			casbinPolicy := ctx.Value(CasbinPolicyContextKey)

			enforcer := stdcasbin.NewEnforcer(casbinModel, casbinPolicy)
			ctx = context.WithValue(ctx, CasbinEnforcerContextKey, enforcer)
			if enforcer.Enforce(subject, object, action) == false {
				return nil, ErrUnauthorized
			}
			return next(ctx, request)
		}
	}
}
