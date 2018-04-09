// Package launchdarkly provides feature flags based on the
// LaunchDarkly services.
package launchdarkly

import (
	"context"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"math"
	"testing"
	"time"

	"github.com/go-kit/kit/log"

	ld "gopkg.in/launchdarkly/go-client.v3"
)

func TestContextEmpty(t *testing.T) {
	user1 := ldUser(context.Background())
	if user1.Key == nil || *user1.Key == "" {
		t.Errorf("expected key: %v\n", user1)
	}
	if !*user1.Anonymous {
		t.Errorf("expected anonymous: %v\n", user1)
	}

	user2 := ldUser(context.Background())
	if *user1.Key == *user2.Key {
		t.Errorf("expected unique random keys: %q vs %q\n", *user1.Key, *user2.Key)
	}
}

func TestContextAssigned(t *testing.T) {
	key := fmt.Sprintf("ld-user-key-%d", time.Now().UnixNano())
	anon := (time.Now().Unix()%2 == 0)
	user := ld.User{Key: &key, Anonymous: &anon}

	ctx := WithUser(context.Background(), user)
	result := ldUser(ctx)

	if *result.Key != key {
		t.Errorf("expected key %q: %v\n", key, result)
	}

	if *result.Anonymous != anon {
		t.Errorf("expected anon %t: %v\n", anon, result)
	}
}

func buildClient(t *testing.T) *ld.LDClient {
	t.Helper()

	cfg := ld.DefaultConfig
	cfg.Offline = true
	cfg.Logger = stdlog.New(ioutil.Discard, "", 0)
	client, _ := ld.MakeCustomClient("", cfg, 0)
	return client
}

func TestBooler(t *testing.T) {
	dflt := (time.Now().Unix()%2 == 0)
	ff := NewBooler(buildClient(t), "test-bool", dflt, log.NewNopLogger())
	val := ff.Bool(context.Background())
	if val != dflt {
		t.Errorf("expected %t; got %t\n", dflt, val)
	}
}

func TestInter(t *testing.T) {
	dflt := time.Now().Unix() % 229 // 50th prime
	ff := NewInter(buildClient(t), "test-int", int(dflt), log.NewNopLogger())
	val := ff.Int(context.Background())
	if val != dflt {
		t.Errorf("expected %d; got %d\n", dflt, val)
	}
}

func TestFloater(t *testing.T) {
	dflt := (float64(time.Now().Unix()%229) * float64(math.Pi))
	ff := NewFloater(buildClient(t), "test-float", dflt, log.NewNopLogger())
	val := ff.Float(context.Background())
	if val != dflt {
		t.Errorf("expected %f; got %f\n", dflt, val)
	}
}

func TestStringer(t *testing.T) {
	dflt := fmt.Sprintf("value %d", time.Now().UnixNano())
	ff := NewStringer(buildClient(t), "test-string", dflt, log.NewNopLogger())
	val := ff.String(context.Background())
	if val != dflt {
		t.Errorf("expected %q; got %q\n", dflt, val)
	}
}
