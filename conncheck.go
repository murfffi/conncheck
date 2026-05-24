package conncheck

import (
	"crypto/tls"
	"errors"
	"net"
	"syscall"
)

// Status represents the known status of the connection
type Status int

const (
	// StatusUnknown means we couldn't determine if the connection is open because
	// of an unexpected error or an unsupported type of connection.
	// Clients should assume the connection is *open*.
	StatusUnknown Status = iota

	// StatusOpen means the connection is open.
	StatusOpen

	// StatusNotOpen means the connection is closed or broken.
	StatusNotOpen
)

// Do checks if the connection is open. This is the primary function of the library.
// It communicates with the network stack but doesn't cause outbound traffic.
// The call completes in microseconds on Unix and up to 1.5 milliseconds on Windows.
// It doesn't interfere with other goroutines using the same connection.
// Supports TCP, TLS, and UDP connections.
func Do(conn net.Conn) Status {
	if tlsConn, ok := conn.(*tls.Conn); ok {
		// We will only peek from the underlying connection, so we don't corrupt the TLS session.
		conn = tlsConn.NetConn()
	}

	sc, ok := conn.(syscall.Conn)
	if !ok {
		return StatusUnknown
	}

	rawConn, err := sc.SyscallConn()
	if err != nil || rawConn == nil {
		if errors.Is(err, syscall.EINVAL) || errors.Is(err, net.ErrClosed) {
			return StatusNotOpen
		}
		return StatusUnknown
	}

	return tryPeek(rawConn)
}
