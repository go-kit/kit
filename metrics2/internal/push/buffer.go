package push

// Buffer produces metrics that are reported on a per-observation basis.
// That is, no statistical aggregation shall occur in the client.
// Several instrumentation systems are built this way, e.g. StatsD.
type Buffer struct {
	prefix string
	addc   chan Add
	setc   chan Set
	obvc   chan Obv
	getc   chan chan collection
	cache  collection
}

// NewBuffer returns a new Buffer ready for use. The prefix is applied to the
// names of all created metrics. The bufSz is the channel buffer depth for each
// metric type, and should be used to prevent blocking in the observation path.
func NewBuffer(prefix string, bufSz int) *Buffer {
	b := &Buffer{
		prefix: prefix,
		addc:   make(chan Add, bufSz),
		setc:   make(chan Set, bufSz),
		obvc:   make(chan Obv, bufSz),
		getc:   make(chan chan collection),
	}
	go b.loop()
	return b
}

// NewCounter returns a forwarding Counter metric.
func (b *Buffer) NewCounter(name string, sampleRate float64) *Counter {
	return NewCounter(b.prefix+name, sampleRate, b.addc)
}

// NewGauge returns a forwarding Gauge metric.
func (b *Buffer) NewGauge(name string) *Gauge {
	return NewGauge(b.prefix+name, b.setc)
}

// NewHistogram returns a forwarding Histogram metric.
func (b *Buffer) NewHistogram(name string, sampleRate float64) *Histogram {
	return NewHistogram(b.prefix+name, sampleRate, b.obvc)
}

func (b *Buffer) loop() {
	for {
		select {
		case a := <-b.addc:
			b.cache.a = append(b.cache.a, a)
		case s := <-b.setc:
			b.cache.s = append(b.cache.s, s)
		case o := <-b.obvc:
			b.cache.o = append(b.cache.o, o)
		case c := <-b.getc:
			c <- b.cache
			b.cache = collection{}
		}
	}
}

// Get returns ordered sets of observations from all counters, gauges, and
// metrics since the last call. Get is destructive, so make sure those
// observations get to their destination!
func (b *Buffer) Get() ([]Add, []Set, []Obv) {
	c := make(chan collection)
	b.getc <- c
	res := <-c
	return res.a, res.s, res.o
}

type collection struct {
	a []Add
	s []Set
	o []Obv
}
