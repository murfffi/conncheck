package conncheck

import (
	"crypto/tls"
	"errors"
	"net"
	"syscall"
	"time"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusOpen
	StatusNotOpen
)

func Do(conn net.Conn) Status {
	if tlsConn, ok := conn.(*tls.Conn); ok {
		// We will only peek from the underlying connection, so we don't corrupt the TLS session.
		conn = tlsConn.NetConn()
	}

	sc, ok := conn.(syscall.Conn)
	if !ok {
		// This happens on WASM
		return StatusUnknown
	}

	// we need to reset the deadline for the checks to work
	err := conn.SetReadDeadline(time.Time{})
	if err != nil {
		// only happens if closing
		return StatusNotOpen
	}

	rawConn, err := sc.SyscallConn()
	if err != nil {
		if errors.Is(err, syscall.EINVAL) {
			return StatusNotOpen
		}
		return StatusUnknown
	}

	return tryPeek(rawConn)
}
