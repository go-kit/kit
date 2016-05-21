package conn

import (
	"net"
	"time"

	"github.com/go-kit/kit/log"
)

// Dialer imitates net.Dial. Dialer is assumed to yield connections that are
// safe for use by multiple concurrent goroutines.
type Dialer func(network, address string) (net.Conn, error)

// AfterFunc imitates time.After.
type AfterFunc func(time.Duration) <-chan time.Time

// Manager manages a net.Conn.
//
// Clients provide a way to create the connection with a Dialer, network, and
// address. Clients should Take the connection when they want to use it, and Put
// back whatever error they receive from its use. When a non-nil error is Put,
// the connection is invalidated, and a new connection is established.
// Connection failures are retried after an exponential backoff.
type Manager struct {
	dialer  Dialer
	network string
	address string
	after   AfterFunc
	logger  log.Logger

	takec chan net.Conn
	putc  chan error
}

// NewManager returns a connection manager using the passed Dialer, network, and
// address. The AfterFunc is used to control exponential backoff and retries.
// For normal use, pass net.Dial and time.After as the Dialer and AfterFunc
// respectively. The logger is used to log errors; pass a log.NopLogger if you
// don't care to receive them.
func NewManager(d Dialer, network, address string, after AfterFunc, logger log.Logger) *Manager {
	m := &Manager{
		dialer:  d,
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

// Take yields the current connection. It may be nil.
func (m *Manager) Take() net.Conn {
	return <-m.takec
}

// Put accepts an error that came from a previously yielded connection. If the
// error is non-nil, the manager will invalidate the current connection and try
// to reconnect, with exponential backoff. Putting a nil error is a no-op.
func (m *Manager) Put(err error) {
	m.putc <- err
}

func (m *Manager) loop() {
	var (
		conn       = dial(m.dialer, m.network, m.address, m.logger) // may block slightly
		connc      = make(chan net.Conn)
		reconnectc <-chan time.Time // initially nil
		backoff    = time.Second
	)

	for {
		select {
		case <-reconnectc:
			reconnectc = nil // one-shot
			go func() { connc <- dial(m.dialer, m.network, m.address, m.logger) }()

		case conn = <-connc:
			if conn == nil {
				// didn't work
				backoff = exponential(backoff) // wait longer
				reconnectc = m.after(backoff)  // try again
			} else {
				// worked!
				backoff = time.Second // reset wait time
				reconnectc = nil      // no retry necessary
			}

		case m.takec <- conn:

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
		conn = nil // just to be sure
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
