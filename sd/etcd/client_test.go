package etcd

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestNewClient(t *testing.T) {
	ClientOptions := ClientOptions{
		Cert:                    "",
		Key:                     "",
		CaCert:                  "",
		DialTimeout:             (2 * time.Second),
		DialKeepAlive:           (2 * time.Second),
		HeaderTimeoutPerRequest: (2 * time.Second),
	}

	client, err := NewClient(
		context.Background(),
		[]string{"http://irrelevant:12345"},
		ClientOptions,
	)
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	if client == nil {
		t.Fatal("expected new Client, got nil")
	}
}

func TestOptions(t *testing.T) {
	//creating new client should fail when providing invalid or missing endpoints
	a, err := NewClient(
		context.Background(),
		[]string{},
		ClientOptions{
			Cert:                    "",
			Key:                     "",
			CaCert:                  "",
			DialTimeout:             (2 * time.Second),
			DialKeepAlive:           (2 * time.Second),
			HeaderTimeoutPerRequest: (2 * time.Second),
		})

	if err == nil {
		t.Errorf("expected error: %v", err)
	}
	if a != nil {
		t.Fatalf("expected client to be nil on failure")
	}

	//creating new client should fail when providing invalid or missing endpoints
	_, err = NewClient(
		context.Background(),
		[]string{"http://irrelevant:12345"},
		ClientOptions{
			Cert:                    "blank.crt",
			Key:                     "blank.key",
			CaCert:                  "blank.cacert",
			DialTimeout:             (2 * time.Second),
			DialKeepAlive:           (2 * time.Second),
			HeaderTimeoutPerRequest: (2 * time.Second),
		})

	if err == nil {
		t.Errorf("expected error: %v", err)
	}
}
