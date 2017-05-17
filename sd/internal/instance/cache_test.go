package instance

import (
	"sync"
	"testing"

	"github.com/go-kit/kit/sd"
)

var _ sd.Instancer = &Cache{} // API check

func TestCache(t *testing.T) {
	// TODO this test is not finished yet

	c := NewCache()

	{
		state := c.State()
		if want, have := 0, len(state.Instances); want != have {
			t.Fatalf("want %v instances, have %v", want, have)
		}
	}

	notification1 := sd.Event{Instances: []string{"x", "y"}}
	notification2 := sd.Event{Instances: []string{"a", "b", "c"}}

	c.Update(notification1)

	// times 2 because we have two observers
	expectedInstances := 2 * (len(notification1.Instances) + len(notification2.Instances))

	wg := sync.WaitGroup{}
	wg.Add(expectedInstances)

	receiver := func(ch chan sd.Event) {
		for state := range ch {
			// count total number of instances received
			for range state.Instances {
				wg.Done()
			}
		}
	}

	f1 := make(chan sd.Event)
	f2 := make(chan sd.Event)
	go receiver(f1)
	go receiver(f2)

	c.Register(f1)
	c.Register(f2)

	c.Update(notification1)
	c.Update(notification2)

	// if state := c.State(); instances == nil {
	// 	if want, have := len(notification2), len(instances); want != have {
	// 		t.Errorf("want length %v, have %v", want, have)
	// 	} else {
	// 		for i := range notification2 {
	// 			if want, have := notification2[i], instances[i]; want != have {
	// 				t.Errorf("want instance %v, have %v", want, have)
	// 			}
	// 		}
	// 	}
	// }

	close(f1)
	close(f2)

	wg.Wait()

	// d.Deregister(f1)

	// d.Unregister(f2)
	// if want, have := 0, len(d.observers); want != have {
	// 	t.Fatalf("want %v observers, have %v", want, have)
	// }
}
