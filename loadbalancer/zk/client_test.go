package zk

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	stdzk "github.com/samuel/go-zookeeper/zk"
)

type logWriter struct {
	t *testing.T
	p string
}

func (lw logWriter) Write(b []byte) (int, error) {
	lw.t.Logf("%s%s", lw.p, string(b))
	return len(b), nil
}

func TestNewClient(t *testing.T) {
	var (
		acl            = stdzk.WorldACL(stdzk.PermRead)
		connectTimeout = 3 * time.Second
		sessionTimeout = 20 * time.Second
		payload        = [][]byte{[]byte("Payload"), []byte("Test")}
	)

	ts, host := startCluster(t)
	defer ts.Stop()

	c, err := NewClient(
		host,
		logger,
		WithACL(acl),
		WithConnectTimeout(connectTimeout),
		WithSessionTimeout(sessionTimeout),
		WithPayload(payload),
	)
	if err != nil {
		t.Fatal(err)
	}
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
}

func TestCreateParentNodes(t *testing.T) {
	ts, host := startCluster(t)
	defer ts.Stop()

	payload := [][]byte{[]byte("Payload"), []byte("Test")}
	client, err := NewClient(host, logger, WithPayload(payload))
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Error("expected pointer to client, got nil")
	}

	p, err := NewPublisher(client, path, NewFactory(""), logger)
	if err != nil {
		t.Fatal(err)
	}
	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}

	zk1, err := ts.Connect(0)
	if err != nil {
		t.Fatalf("Connect returned error: %+v", err)
	}
	data, _, err := zk1.Get(path)
	if err != nil {
		t.Fatal(err)
	}
	// test Client implementation of CreateParentNodes. It should have created
	// our payload
	if bytes.Compare(data, payload[1]) != 0 {
		t.Errorf("want %q, have %q", payload[1], data)
	}

}

func TestGetEntries(t *testing.T) {
	var instancePayload = "protocol://hostname:port/routing"
	ts, host := startCluster(t)
	defer ts.Stop()

	zk1, err := ts.Connect(0)
	if err != nil {
		t.Fatalf("Connect returned error: %+v", err)
	}
	defer zk1.Close()

	c, err := NewClient(host, logger)
	p, err := NewPublisher(c, path, NewFactory(""), logger)
	if err != nil {
		t.Fatal(err)
	}

	_, err = zk1.Create(
		path+"/instance1",
		[]byte(instancePayload),
		stdzk.FlagEphemeral|stdzk.FlagSequence,
		stdzk.WorldACL(stdzk.PermAll),
	)
	if err != nil {
		t.Fatalf("Unable to create test ephemeral znode: %+v", err)
	}

	time.Sleep(1 * time.Second)

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 1, len(endpoints); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func startCluster(t *testing.T) (*stdzk.TestCluster, []string) {
	// Start ZooKeeper Test Cluster
	ts, err := stdzk.StartTestCluster(1, nil, logWriter{t: t, p: "[ZKERR] "})
	if err != nil {
		t.Fatal(err)
	}
	host := fmt.Sprintf("localhost:%d", ts.Servers[0].Port)
	return ts, []string{host}
}
