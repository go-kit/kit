package zk

import (
	"bytes"
	"testing"
	"time"

	stdzk "github.com/samuel/go-zookeeper/zk"
)

func TestNewClient(t *testing.T) {
	var (
		acl            = stdzk.WorldACL(stdzk.PermRead)
		connectTimeout = 3 * time.Second
		sessionTimeout = 20 * time.Second
		payload        = [][]byte{[]byte("Payload"), []byte("Test")}
	)

	c, err := NewClient(
		[]string{"FailThisInvalidHost!!!"},
		logger,
	)

	time.Sleep(1 * time.Millisecond)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	calledEventHandler := false
	eventHandler := func(event stdzk.Event) {
		calledEventHandler = true
	}
	c, err = NewClient(
		[]string{"localhost"},
		logger,
		ACL(acl),
		ConnectTimeout(connectTimeout),
		SessionTimeout(sessionTimeout),
		Payload(payload),
		EventHandler(eventHandler),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Stop()
	clientImpl, ok := c.(*client)
	if !ok {
		t.Errorf("retrieved incorrect Client implementation")
	}
	if want, have := acl, clientImpl.acl; want[0] != have[0] {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := connectTimeout, clientImpl.connectTimeout; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := sessionTimeout, clientImpl.sessionTimeout; want != have {
		t.Errorf("want %q, have %q", want, have)
	}
	if want, have := payload, clientImpl.rootNodePayload; bytes.Compare(want[0], have[0]) != 0 || bytes.Compare(want[1], have[1]) != 0 {
		t.Errorf("want %q, have %q", want, have)
	}
	// Allow EventHandler to be called
	time.Sleep(1 * time.Millisecond)

	if want, have := true, calledEventHandler; want != have {
		t.Errorf("want %t, have %t", want, have)
	}
}

func TestOptions(t *testing.T) {
	_, err := NewClient([]string{"localhost"}, logger, Credentials("valid", "credentials"))
	if err != nil && err != stdzk.ErrNoServer {
		t.Errorf("unexpected error: %q", err)
	}

	_, err = NewClient([]string{"localhost"}, logger, Credentials("nopass", ""))
	if want, have := err, ErrInvalidCredentials; want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	_, err = NewClient([]string{"localhost"}, logger, ConnectTimeout(0))
	if err == nil {
		t.Errorf("expected connect timeout error")
	}

	_, err = NewClient([]string{"localhost"}, logger, SessionTimeout(0))
	if err == nil {
		t.Errorf("expected connect timeout error")
	}
}
