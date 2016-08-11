package dogstatsd

import "sync"

type rateMap struct {
	mtx sync.RWMutex
	m   map[string]float64
}

func newRateMap() *rateMap {
	return &rateMap{
		m: map[string]float64{},
	}
}

func (m *rateMap) set(name string, rate float64) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.m[name] = rate
}

func (m *rateMap) get(name string) float64 {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	f, ok := m.m[name]
	if !ok {
		f = 1.0
	}
	return f
}
