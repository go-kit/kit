package dogstatsd

import (
	"io/ioutil"
	"time"

	"github.com/go-kit/kit/log"
)

func ExampleDogstatsd_WriteLoop() {
	d := New("foo_service", 1024, log.NewNopLogger())
	t := time.NewTicker(time.Second)
	defer t.Stop()
	go d.WriteLoop(t.C, ioutil.Discard)

	// Now, use d to create counters, gauges, and histograms.
	// Pass those metrics to the components that will use them.
}

func ExampleDogstatsd_SendLoop() {
	d := New("bar_service", 1024, log.NewNopLogger())
	t := time.NewTicker(time.Second)
	defer t.Stop()
	go d.SendLoop(t.C, "udp", "dogstatsd.internal:8125")

	// Now, use d to create counters, gauges, and histograms.
	// Pass those metrics to the components that will use them.
}
