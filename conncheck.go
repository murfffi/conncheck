package conncheck

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
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
	res := strconv.Itoa(int(StatusUnknown))
	switch s {
	case StatusUnknown:
		return res + ": StatusUnknown"
	case StatusOpen:
		return res + ": StatusOpen"
	case StatusNotOpen:
		return res + ": StatusNotOpen"
	default:
		return res + ": unknown"
	}
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
	if err == nil {
		err = tryPeek(rawConn)
	} // else means that the system handle is not set

	if err != nil {
		return StatusNotOpen
	}
	return StatusOpen
}

func IsBroken(conn net.Conn) bool {
	return Do(conn) == StatusNotOpen
}
