package conn

import (
	"net"
	"time"

	"github.com/go-kit/kit/log"
)

// Dialer dials a network and address. net.Dial is a good default Dialer.
type Dialer func(network, address string) (net.Conn, error)

// AfterFunc imitates time.After.
type AfterFunc func(time.Duration) <-chan time.Time

// Manager manages a net.Conn. Clients should take the conn when they want to
// use it, and put back whatever error they receive from an e.g. Write. When a
// non-nil error is put, the conn is invalidated and a new conn is established.
// Connection failures are retried after an exponential backoff.
type Manager struct {
	dial    Dialer
	network string
	address string
	after   AfterFunc
	logger  log.Logger

	takec chan net.Conn
	putc  chan error
}

func NewManager(d Dialer, network, address string, after AfterFunc, logger log.Logger) *Manager {
	m := &Manager{
		dial:    d,
		network: network,
		address: address,
		after:   after,
		logger:  logger,

		takec: make(chan net.Conn),
		putc:  make(chan error),
	}
	go m.loop()
	return m
}

func (m *Manager) Take() net.Conn {
	return <-m.takec
}

func (m *Manager) Put(err error) {
	m.putc <- err
}

func (m *Manager) loop() {
	var (
		conn       = dial(m.dial, m.network, m.address, m.logger) // may block slightly
		connc      = make(chan net.Conn)
		reconnectc <-chan time.Time // initially nil
		backoff    = time.Second
	)

	for {
		select {
		case <-reconnectc:
			reconnectc = nil
			go func() { connc <- dial(m.dial, m.network, m.address, m.logger) }()

		case conn = <-connc:
			if conn == nil {
				backoff = exponential(backoff)
				reconnectc = m.after(backoff)
			} else {
				backoff = time.Second
				reconnectc = nil
			}

		case m.takec <- conn:
			// might be nil

		case err := <-m.putc:
			if err != nil && conn != nil {
				m.logger.Log("err", err)
				conn = nil                            // connection is bad
				reconnectc = m.after(time.Nanosecond) // trigger immediately
			}
		}
	}
}

func dial(d Dialer, network, address string, logger log.Logger) net.Conn {
	conn, err := d(network, address)
	if err != nil {
		logger.Log("err", err)
		conn = nil
	}
	return conn
}

func exponential(d time.Duration) time.Duration {
	d *= 2
	if d > time.Minute {
		d = time.Minute
	}
	return d
}
