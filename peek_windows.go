//go:build windows

package conncheck

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

const MSG_PEEK uint32 = windows.MSG_PEEK

func tryPeek(rawConn syscall.RawConn) Status {
	var n uint32

	var recvErr, sockOptErr error
	readErr := rawConn.Read(func(fd uintptr) bool {
		h := windows.Handle(fd)

		var oldTimeout int
		oldTimeout, sockOptErr = windows.GetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO)
		if sockOptErr != nil {
			return true
		}

		sockOptErr = windows.SetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO, 1 /* millis */)
		if sockOptErr != nil {
			return true
		}
		buf := [1]byte{}
		wsabuf := windows.WSABuf{
			Len: uint32(len(buf)),
			Buf: &buf[0],
		}

		recvErr = windows.WSARecv(h, &wsabuf, 1, &n, new(MSG_PEEK), nil, nil)
		sockOptErr = windows.SetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO, oldTimeout)
		return true
	})

	if sockOptErr != nil {
		return StatusUnknown
	}

	if readErr != nil {
		return StatusNotOpen
	}

	if n > 0 {
		// connection is open and there is something in the buffer
		// recvErr should be nil here since we read something, but if it isn't,
		// return it as processErr to invalidate the result
		if recvErr != nil {
			return StatusUnknown
		}
		return StatusOpen
	}

	if errors.Is(recvErr, windows.WSAETIMEDOUT) || // if the connection is blocking
		errors.Is(recvErr, windows.WSAEWOULDBLOCK) || // if the connection is non-blocking (unlikely)
		// based on example in golang/go/blob/364de84f/src/internal/poll/fd_windows.go#L1262
		errors.Is(recvErr, windows.WSAEMSGSIZE) {
		// connection is open and there is nothing in the buffer
		return StatusOpen
	}

	return StatusNotOpen // recvErr is not nil or n == 0 (EOF)
}
