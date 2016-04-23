package etcd_test

import (
	etcd "github.com/coreos/etcd/client"
	kitetcd "github.com/go-kit/kit/loadbalancer/etcd"
	"golang.org/x/net/context"
)

import "testing"

type MockKeysAPI struct {
	getErr   error
	watchErr error
}

func (mock *MockKeysAPI) Get() (*etcd.Response, error) {
	return &etcd.Response{}, mock.getErr
}

func TestNoCertificateClient(t *testing.T) {
	context := context.Background()

	_, err := kitetcd.NewClient(context, []string{"http://localhost:2379"}, nil)
	if err != nil {
		t.Fatalf("failed to create new client %v", err)
	}
}

func TestNoMachines(t *testing.T) {
	context := context.Background()

	_, err := kitetcd.NewClient(context, []string{}, nil)
	if err == nil {
		t.Fatalf("should return error if no machines provided")
	}
}

func TestWrongCert(t *testing.T) {
	context := context.Background()
	ops := &kitetcd.ClientOptions{
		Key:    "./test_certs/host.key",
		Cert:   "boom",
		CaCert: "./test_certs/rootCA.crt",
	}
	_, err := kitetcd.NewClient(context, []string{"http://localhost:2379"}, ops)
	if err == nil {
		t.Fatalf("should return an error if Certificate is corrupted")
	}
}

func TestWrongCaCert(t *testing.T) {
	context := context.Background()
	ops := &kitetcd.ClientOptions{
		Key:    "./test_certs/host.key",
		Cert:   "./test_certs/host.crt",
		CaCert: "boom",
	}
	_, err := kitetcd.NewClient(context, []string{"http://localhost:2379"}, ops)
	if err == nil {
		t.Fatalf("should return an error if CA Certificate is corrupted")
	}
}

func TestNewClient(t *testing.T) {
	context := context.Background()
	ops := &kitetcd.ClientOptions{
		Key:    "./test_certs/host.key",
		Cert:   "./test_certs/host.crt",
		CaCert: "./test_certs/rootCA.crt",
	}
	c, err := kitetcd.NewClient(context, []string{"http://localhost:2379"}, ops)
	if err != nil {
		t.Fatalf("failed to create new client %v", err)
	}
	_, err = c.GetEntries("key")
	if err.Error() != etcd.ErrClusterUnavailable.Error() {
		t.Fatalf("unexpected error %v", err)
	}
	ch := make(chan *etcd.Response)
	// should fail but no idea how to check it
	c.WatchPrefix("foo", ch)
}
