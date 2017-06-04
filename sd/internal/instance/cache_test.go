package instance

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = &Cache{} // API check

// The test verifies the following:
//   registering causes initial notification of the current state
//   notifications delivered to two receivers
//   identical notifications cause no updates
//   different update causes new notification
//   instances are sorted
//   no updates after de-registering
func TestCache(t *testing.T) {
	e1 := sd.Event{Instances: []string{"y", "x"}} // not sorted
	e2 := sd.Event{Instances: []string{"c", "a", "b"}}

	c := NewCache()
	if want, have := 0, len(c.State().Instances); want != have {
		t.Fatalf("want %v instances, have %v", want, have)
	}

	c.Update(e1) // sets initial state
	if want, have := 2, len(c.State().Instances); want != have {
		t.Fatalf("want %v instances, have %v", want, have)
	}

	r1 := make(chan sd.Event)
	go c.Register(r1)
	expectUpdate(t, r1, []string{"x", "y"})

	r2 := make(chan sd.Event)
	go c.Register(r2)
	expectUpdate(t, r2, []string{"x", "y"})

	// send the same instances but in different order.
	// because it's a duplicate it should not cause new notification.
	// if it did, this call would deadlock trying to send to channels with no readers
	c.Update(sd.Event{Instances: []string{"x", "y"}})
	expectNoUpdate(t, r1)
	expectNoUpdate(t, r2)

	go c.Update(e2) // different set
	expectUpdate(t, r1, []string{"a", "b", "c"})
	expectUpdate(t, r2, []string{"a", "b", "c"})

	c.Deregister(r1)
	c.Deregister(r2)
	close(r1)
	close(r2)
	// if deregister didn't work, Update would panic on closed channels
	c.Update(e1)
}

func expectUpdate(t *testing.T, r chan sd.Event, expect []string) {
	select {
	case e := <-r:
		if want, have := expect, e.Instances; !reflect.DeepEqual(want, have) {
			t.Fatalf("want: %v, have: %v", want, have)
		}
	case <-time.After(time.Second):
		t.Fatalf("did not receive expected update")
	}
}

func expectNoUpdate(t *testing.T, r chan sd.Event) {
	select {
	case e := <-r:
		t.Errorf("received unexpected update %v", e)
	case <-time.After(time.Millisecond):
		return // as expected
	}
}
