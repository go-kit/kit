package zk

import (
	"testing"
	"time"
)

func TestPublisher(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, newFactory(""), logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	if _, err := p.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadFactory(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, newFactory("kaboom"), logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	// instance1 came online
	client.AddService(path+"/instance1", "kaboom")

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data")

	if err = asyncTest(100*time.Millisecond, 1, p); err != nil {
		t.Error(err)
	}
}

func TestServiceUpdate(t *testing.T) {
	client := newFakeClient()

	p, err := NewPublisher(client, path, newFactory(""), logger)
	if err != nil {
		t.Fatalf("failed to create new publisher: %v", err)
	}
	defer p.Stop()

	endpoints, err := p.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// instance1 came online
	client.AddService(path+"/instance1", "zookeeper_node_data")

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data2")

	// we should have 2 instances
	if err = asyncTest(100*time.Millisecond, 2, p); err != nil {
		t.Error(err)
	}

	// watch triggers an error...
	client.SendErrorOnWatch()

	// test if error was consumed
	if err = client.ErrorIsConsumed(100 * time.Millisecond); err != nil {
		t.Error(err)
	}

	// instance3 came online
	client.AddService(path+"/instance3", "zookeeper_node_data3")

	// we should have 3 instances
	if err = asyncTest(100*time.Millisecond, 3, p); err != nil {
		t.Error(err)
	}

	// instance1 goes offline
	client.RemoveService(path + "/instance1")

	// instance2 goes offline
	client.RemoveService(path + "/instance2")

	// we should have 1 instance
	if err = asyncTest(100*time.Millisecond, 1, p); err != nil {
		t.Error(err)
	}
}

func TestBadPublisherCreate(t *testing.T) {
	client := newFakeClient()
	client.SendErrorOnWatch()
	p, err := NewPublisher(client, path, newFactory(""), logger)
	if err == nil {
		t.Error("expected error on new publisher")
	}
	if p != nil {
		t.Error("expected publisher not to be created")
	}
	p, err = NewPublisher(client, "BadPath", newFactory(""), logger)
	if err == nil {
		t.Error("expected error on new publisher")
	}
	if p != nil {
		t.Error("expected publisher not to be created")
	}
}
