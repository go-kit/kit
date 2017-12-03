package backoff

import (
	"testing"
	"time"
)

func TestNext(t *testing.T) {
	b := ExponentialBackoff{}
	b.currentInterval.Store(time.Duration(12))

	next := b.next()

	if next < 12 || next > 36 {
		t.Errorf("Expected next to be between 12 and 36, got %d", 12)
	}
}

func TestNextBackoffMax(t *testing.T) {
	max := time.Duration(13)
	b := ExponentialBackoff{
		Max: max,
	}
	b.currentInterval.Store(time.Duration(14))
	next := b.NextBackoff()
	if next != max {
		t.Errorf("Expected next to be max, %d, but got %d", max, next)
	}

	current := b.currentInterval.Load().(time.Duration)
	if current != max {
		t.Errorf("Expected currentInterval to be max, %d, but got %d", max, current)
	}
}
