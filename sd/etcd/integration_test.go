// +build integration

package etcd

import (
	"flag"
	"kit/log"
	"os"
	"testing"
	"time"

	etcdc "github.com/coreos/etcd/client"
	etcdi "github.com/coreos/etcd/integration"
	"golang.org/x/net/context"
)

var (
	host             []string
	kitClientOptions ClientOptions
)

func TestMain(m *testing.M) {
	flag.Parse()

	kitClientOptions = ClientOptions{
		Cert:                    "",
		Key:                     "",
		CaCert:                  "",
		DialTimeout:             (2 * time.Second),
		DialKeepAlive:           (2 * time.Second),
		HeaderTimeoutPerRequest: (2 * time.Second),
	}

	code := m.Run()

	os.Exit(code)
}

func TestRegistrar(t *testing.T) {
	ts := etcdi.NewCluster(t, 1)
	ts.Launch(t)
	kitClient, err := NewClient(context.Background(), []string{ts.URL(0)}, kitClientOptions)

	// Valid registrar should pass
	registrar := NewRegistrar(kitClient, Service{
		Key:   "somekey",
		Value: "somevalue",
		DeleteOptions: &etcdc.DeleteOptions{
			PrevValue: "",
			PrevIndex: 0,
			Recursive: true,
			Dir:       false,
		},
	}, log.NewNopLogger())

	registrar.Register()
	r1, err := kitClient.GetEntries(registrar.service.Key)
	if err != nil {
		t.Fatalf("unexpected error when getting value for deregistered key: %v", err)
	}

	if want, have := registrar.service.Value, r1[0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	registrar.Deregister()
	r2, err := kitClient.GetEntries(registrar.service.Key)
	if len(r2) > 0 {
		t.Fatalf("unexpected value found for deregistered key: %s", r2)
	}

	// Registrar with no key should register but value will be blank
	registrarNoKey := NewRegistrar(kitClient, Service{
		Key:   "",
		Value: "somevalue",
		DeleteOptions: &etcdc.DeleteOptions{
			PrevValue: "",
			PrevIndex: 0,
			Recursive: true,
			Dir:       false,
		},
	}, log.NewNopLogger())

	registrarNoKey.Register()
	r3, err := kitClient.GetEntries(registrarNoKey.service.Key)
	if err != nil {
		t.Errorf("unexpected error when getting value for entry with no key: %v", err)
	}

	if want, have := "", r3[0]; want != have {
		t.Fatalf("want %q, have %q", want, have)
	}

	// Registrar with no value should not register anything
	registrarNoValue := NewRegistrar(kitClient, Service{
		Key:   "somekey",
		Value: "",
		DeleteOptions: &etcdc.DeleteOptions{
			PrevValue: "",
			PrevIndex: 0,
			Recursive: true,
			Dir:       false,
		},
	}, log.NewNopLogger())

	registrarNoValue.Register()
	r4, err := kitClient.GetEntries(registrarNoValue.service.Key)
	if err == nil {
		t.Errorf("expected error when getting value for entry key which attempted to register with no value")
	}

	if len(r4) > 0 {
		t.Fatalf("unexpected value retreived when getting value for entry with no value")
	}

	ts.Terminate(t)
}
