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

		// RecvFrom is a simpler alternative of WSARecv, but it sometimes returns
		// "WSAEAFNOSUPPORT - Address family not supported by protocol family", for some reason.
		recvErr = windows.WSARecv(h, &wsabuf, 1, &n, new(MSG_PEEK), nil, nil)

		sockOptErr = windows.SetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO, oldTimeout)
		// It should not be possible to get an error here if the connection is open.
		// All documented failure reasons in https://learn.microsoft.com/en-us/windows/win32/api/winsock/nf-winsock-setsockopt
		// mean that either the socket is closed or that the option is not supported, which would have failed earlier.
		return true
	})

	if sockOptErr != nil {
		return StatusUnknown
	}

	if readErr != nil {
		return StatusNotOpen
	}

	if n > 0 {
		// NB: recvErr may not be nil even with n > 0 e.g. WSAEMSGSIZE.
		// Still, if we read something, the connection is open.
		return StatusOpen
	}

	if errors.Is(recvErr, windows.WSAETIMEDOUT) || // if the connection is in blocking mode (typical)
		errors.Is(recvErr, windows.WSAEWOULDBLOCK) || // if the connection is in non-blocking mode
		// based on example in golang/go/blob/364de84f/src/internal/poll/fd_windows.go#L1262
		errors.Is(recvErr, windows.WSAEMSGSIZE) {
		// connection is open and there is nothing in the buffer
		return StatusOpen
	}

	return StatusNotOpen // recvErr is not nil or n == 0 (EOF)
}
