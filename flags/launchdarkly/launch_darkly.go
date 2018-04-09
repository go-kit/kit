// Package launchdarkly provides feature flags based on the
// LaunchDarkly services.
package launchdarkly

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/go-kit/kit/flags"
	"github.com/go-kit/kit/log"

	ld "gopkg.in/launchdarkly/go-client.v3"
)

// contextKey type is unexported, unique to this package
type contextKey int

// userKey is what marks the LaunchDarkly User struct in the context
const userKey contextKey = 0

// WithUser enables clients to set the User object that will be used to
// calculate the feature flag output.
func WithUser(ctx context.Context, user ld.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// ldUser finds the assigned User struct from the context, or returns
// an anonymous user with a random (16-byte) key
func ldUser(ctx context.Context) ld.User {
	user, ok := ctx.Value(userKey).(ld.User)
	if ok {
		return user
	}
	key := make([]byte, 16)
	io.ReadFull(rand.Reader, key)
	return ld.NewAnonymousUser(fmt.Sprintf("%x", key))
}

// NewBooler builds a Booler that returns a boolean, as per the configured
// LaunchDarkly client
func NewBooler(client *ld.LDClient, key string, defaultVal bool, l log.Logger) flags.Booler {
	return flags.BoolerFunc(func(ctx context.Context) bool {
		user := ldUser(ctx)
		val, err := client.BoolVariation(key, user, defaultVal)
		if err != nil {
			l.Log("launchdarkly bool", err.Error(), "key", key)
		}
		return val
	})
}

// NewInter builds an Inter that returns an int64, as per the configured
// LaunchDarkly client
func NewInter(client *ld.LDClient, key string, defaultVal int, l log.Logger) flags.Inter {
	return flags.InterFunc(func(ctx context.Context) int64 {
		user := ldUser(ctx)
		val, err := client.IntVariation(key, user, defaultVal)
		if err != nil {
			l.Log("launchdarkly int", err.Error(), "key", key)
		}
		return int64(val)
	})
}

// NewFloater builds a Floater that returns a float64, as per the configured
// LaunchDarkly client
func NewFloater(client *ld.LDClient, key string, defaultVal float64, l log.Logger) flags.Floater {
	return flags.FloaterFunc(func(ctx context.Context) float64 {
		user := ldUser(ctx)
		val, err := client.Float64Variation(key, user, defaultVal)
		if err != nil {
			l.Log("launchdarkly float", err.Error(), "key", key)
		}
		return val
	})
}

// NewStringer builds a Stringer that returns a string, as per the configured
// LaunchDarkly client
func NewStringer(client *ld.LDClient, key string, defaultVal string, l log.Logger) flags.Stringer {
	return flags.StringerFunc(func(ctx context.Context) string {
		user := ldUser(ctx)
		val, err := client.StringVariation(key, user, defaultVal)
		if err != nil {
			l.Log("launchdarkly string", err.Error(), "key", key)
		}
		return val
	})
}
