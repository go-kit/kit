// +build integration

package zk

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	stdzk "github.com/samuel/go-zookeeper/zk"
)

var (
	host []string
)

func TestMain(m *testing.M) {
	flag.Parse()

	fmt.Println("Starting ZooKeeper server...")

	ts, err := stdzk.StartTestCluster(1, nil, nil)
	if err != nil {
		fmt.Printf("ZooKeeper server error: %v\n", err)
		os.Exit(1)
	}

	host = []string{fmt.Sprintf("localhost:%d", ts.Servers[0].Port)}
	code := m.Run()

	ts.Stop()
	os.Exit(code)
}

func TestCreateParentNodesOnServer(t *testing.T) {
	payload := [][]byte{[]byte("Payload"), []byte("Test")}
	c1, err := NewClient(host, logger, Payload(payload))
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	if c1 == nil {
		t.Fatal("Expected pointer to client, got nil")
	}
	defer c1.Stop()

	p, err := NewPublisher(c1, path, newFactory(""), logger)
	if err != nil {
		t.Fatalf("Unable to create Publisher: %v", err)
	}
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	c2, err := NewClient(host, logger)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	defer c2.Stop()
	data, _, err := c2.(*client).Get(path)
	if err != nil {
		t.Fatal(err)
	}
	// test Client implementation of CreateParentNodes. It should have created
	// our payload
	if bytes.Compare(data, payload[1]) != 0 {
		t.Errorf("want %s, have %s", payload[1], data)
	}

}

func TestCreateBadParentNodesOnServer(t *testing.T) {
	c, _ := NewClient(host, logger)
	defer c.Stop()

	_, err := NewPublisher(c, "invalid/path", newFactory(""), logger)

	if want, have := stdzk.ErrInvalidPath, err; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestCredentials1(t *testing.T) {
	acl := stdzk.DigestACL(stdzk.PermAll, "user", "secret")
	c, _ := NewClient(host, logger, ACL(acl), Credentials("user", "secret"))
	defer c.Stop()

	_, err := NewPublisher(c, "/acl-issue-test", newFactory(""), logger)

	if err != nil {
		t.Fatal(err)
	}
}

func TestCredentials2(t *testing.T) {
	acl := stdzk.DigestACL(stdzk.PermAll, "user", "secret")
	c, _ := NewClient(host, logger, ACL(acl))
	defer c.Stop()

	_, err := NewPublisher(c, "/acl-issue-test", newFactory(""), logger)

	if err != stdzk.ErrNoAuth {
		t.Errorf("want %v, have %v", stdzk.ErrNoAuth, err)
	}
}

func TestConnection(t *testing.T) {
	c, _ := NewClient(host, logger)
	c.Stop()

	_, err := NewPublisher(c, "/acl-issue-test", newFactory(""), logger)

	if err != ErrClientClosed {
		t.Errorf("want %v, have %v", ErrClientClosed, err)
	}
}

func TestGetEntriesOnServer(t *testing.T) {
	var instancePayload = "protocol://hostname:port/routing"

	c1, err := NewClient(host, logger)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}

	defer c1.Stop()

	c2, err := NewClient(host, logger)
	p, err := NewPublisher(c2, path, newFactory(""), logger)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Stop()

	c2impl, _ := c2.(*client)
	_, err = c2impl.Create(
		path+"/instance1",
		[]byte(instancePayload),
		stdzk.FlagEphemeral|stdzk.FlagSequence,
		stdzk.WorldACL(stdzk.PermAll),
	)
	if err != nil {
		t.Fatalf("Unable to create test ephemeral znode 1: %v", err)
	}
	_, err = c2impl.Create(
		path+"/instance2",
		[]byte(instancePayload+"2"),
		stdzk.FlagEphemeral|stdzk.FlagSequence,
		stdzk.WorldACL(stdzk.PermAll),
	)
	if err != nil {
		t.Fatalf("Unable to create test ephemeral znode 2: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 2, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

func TestGetEntriesPayloadOnServer(t *testing.T) {
	c, err := NewClient(host, logger)
	if err != nil {
		t.Fatalf("Connect returned error: %v", err)
	}
	_, eventc, err := c.GetEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.(*client).Create(
		path+"/instance3",
		[]byte("just some payload"),
		stdzk.FlagEphemeral|stdzk.FlagSequence,
		stdzk.WorldACL(stdzk.PermAll),
	)
	if err != nil {
		t.Fatalf("Unable to create test ephemeral znode: %v", err)
	}
	select {
	case event := <-eventc:
		if want, have := stdzk.EventNodeChildrenChanged.String(), event.Type.String(); want != have {
			t.Errorf("want %s, have %s", want, have)
		}
	case <-time.After(20 * time.Millisecond):
		t.Errorf("expected incoming watch event, timeout occurred")
	}

}
