package conncheck

import (
	"crypto/tls"
	"fmt"
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

func (s Status) String() string {
	var name string
	switch s {
	case StatusUnknown:
		name = "StatusUnknown"
	case StatusOpen:
		name = "StatusOpen"
	case StatusNotOpen:
		name = "StatusNotOpen"
	default:
		name = "unknown"
	}
	return fmt.Sprintf("%d: %s", int(s), name)
}

var _ fmt.Stringer = StatusOpen

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
	} // else means that the system handle is not set

	return tryPeek(rawConn)
}
