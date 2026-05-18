package conncheck

import (
	"crypto/tls"
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
		conn = tlsConn.NetConn()
	}

	sc, ok := conn.(syscall.Conn)
	if !ok {
		return StatusUnknown
	}

	_ = conn.SetReadDeadline(time.Time{})
	rawConn, err := sc.SyscallConn()

	if err != nil {
		return StatusUnknown
	}

	return tryPeek(rawConn)
}
