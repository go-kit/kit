package instance

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = (*Cache)(nil) // API check

// The test verifies the following:
//   registering causes initial notification of the current state
//   instances are sorted
//   different update causes new notification
//   identical notifications cause no updates
//   no updates after de-registering
func TestCache(t *testing.T) {
	e1 := sd.Event{Instances: []string{"y", "x"}} // not sorted
	e2 := sd.Event{Instances: []string{"c", "a", "b"}}

	cache := NewCache()
	if want, have := 0, len(cache.State().Instances); want != have {
		t.Fatalf("want %v instances, have %v", want, have)
	}

	cache.Update(e1) // sets initial state
	if want, have := 2, len(cache.State().Instances); want != have {
		t.Fatalf("want %v instances, have %v", want, have)
	}

	r1 := make(chan sd.Event)
	go cache.Register(r1)
	expectUpdate(t, r1, []string{"x", "y"})

	go cache.Update(e2) // different set
	expectUpdate(t, r1, []string{"a", "b", "c"})

	cache.Deregister(r1)
	close(r1)
}

func expectUpdate(t *testing.T, r chan sd.Event, expect []string) {
	select {
	case e := <-r:
		if want, have := expect, e.Instances; !reflect.DeepEqual(want, have) {
			t.Fatalf("want: %v, have: %v", want, have)
		}
	case <-time.After(time.Second):
		t.Fatalf("did not receive expected update %v", expect)
	}
}

func TestRegistry(t *testing.T) {
	reg := make(registry)
	c1 := make(chan sd.Event, 1)
	c2 := make(chan sd.Event, 1)
	reg.register(c1)
	reg.register(c2)

	// validate that both channels receive the update
	reg.broadcast(sd.Event{Instances: []string{"x", "y"}})
	if want, have := []string{"x", "y"}, (<-c1).Instances; !reflect.DeepEqual(want, have) {
		t.Fatalf("want: %v, have: %v", want, have)
	}
	if want, have := []string{"x", "y"}, (<-c2).Instances; !reflect.DeepEqual(want, have) {
		t.Fatalf("want: %v, have: %v", want, have)
	}

	reg.deregister(c1)
	reg.deregister(c2)
	close(c1)
	close(c2)
	// if deregister didn't work, broadcast would panic on closed channels
	reg.broadcast(sd.Event{Instances: []string{"x", "y"}})
}
