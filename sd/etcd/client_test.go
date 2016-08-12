package etcd

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(
		context.Background(),
		[]string{"http://irrelevant:12345"},
		ClientOptions{
			DialTimeout:             2 * time.Second,
			DialKeepAlive:           2 * time.Second,
			HeaderTimeoutPerRequest: 2 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	if client == nil {
		t.Fatal("expected new Client, got nil")
	}
}

// NewClient should fail when providing invalid or missing endpoints.
func TestOptions(t *testing.T) {
	a, err := NewClient(
		context.Background(),
		[]string{},
		ClientOptions{
			Cert:                    "",
			Key:                     "",
			CACert:                  "",
			DialTimeout:             2 * time.Second,
			DialKeepAlive:           2 * time.Second,
			HeaderTimeoutPerRequest: 2 * time.Second,
		},
	)
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if a != nil {
		t.Fatalf("expected client to be nil on failure")
	}

	_, err = NewClient(
		context.Background(),
		[]string{"http://irrelevant:12345"},
		ClientOptions{
			Cert:                    "blank.crt",
			Key:                     "blank.key",
			CACert:                  "blank.CACert",
			DialTimeout:             2 * time.Second,
			DialKeepAlive:           2 * time.Second,
			HeaderTimeoutPerRequest: 2 * time.Second,
		},
	)
	if err == nil {
		t.Errorf("expected error: %v", err)
	}
}
