//go:build windows

package conncheck

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

func tryPeek(rawConn syscall.RawConn) Status {
	var n uint32

	var recvErr, sockOptErr, sockOptResetErr error

	var peek = func(h windows.Handle) {
		// Unlike Linux, Windows doesn't support MSG_DONTWAIT to do a non-blocking operation regardless of
		// socket mode. Go seems to use sockets in blocking mode on Windows as of 1.26.
		// We also shouldn't change the mode because we can't read the current one, and we wouldn't be able
		// to change it back. Instead, we follow the example of https://github.com/jackc/pgx/pull/1629 and
		// use deadlines/timeouts to avoid blocking for too long.

		// We use the syscall.RawConn so we can peek, not the net.Conn,
		// and thus need to use a timeout instead of a deadline. See SO_RCVTIMEO at
		// https://learn.microsoft.com/en-us/windows/win32/winsock/sol-socket-socket-options

		var oldTimeout int
		oldTimeout, sockOptErr = windows.GetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO)
		if sockOptErr != nil {
			return
		}

		sockOptErr = windows.SetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO, 1 /* millis */)
		if sockOptErr != nil {
			return
		}
		buf := [1]byte{}
		wsabuf := windows.WSABuf{
			Len: uint32(len(buf)),
			Buf: &buf[0],
		}

		flags := uint32(windows.MSG_PEEK)
		// We model this call on the one in go/internal/poll/fd_windows.go#FD.RawRead
		recvErr = windows.WSARecv(h, &wsabuf, 1 /* number of buffers */, &n, &flags, nil, nil)

		sockOptResetErr = windows.SetsockoptInt(h, windows.SOL_SOCKET, windows.SO_RCVTIMEO, oldTimeout)
		// It should not be possible to get an error here if the connection is open.
		// All documented failure reasons in https://learn.microsoft.com/en-us/windows/win32/api/winsock/nf-winsock-setsockopt
		// mean that either the socket is closed or that the option is not supported, which would have failed earlier.
	}
	readErr := rawConn.Read(func(fd uintptr) bool {
		peek(windows.Handle(fd))
		return true // escape out of the RawConn.Read loop
	})

	if sockOptErr != nil {
		// We couldn't set the timeout and do the check.
		return StatusUnknown
	}

	if readErr != nil {
		return StatusNotOpen
	}

	if n > 0 || // we peeked something,
		// or there was nothing in the buffer, which is indicated by n == 0 and either:
		errors.Is(recvErr, windows.WSAETIMEDOUT) || // if the connection is in blocking mode (typical)
		errors.Is(recvErr, windows.WSAEWOULDBLOCK) || // if the connection is in non-blocking mode
		errors.Is(recvErr, windows.WSAEMSGSIZE) { // like in go/internal/poll/fd_windows.go#FD.RawRead

		// connection is open and there is nothing in the buffer
		// recvErr may not be nil even with n > 0. Still, if we read something, the connection is open.

		if sockOptResetErr != nil {
			// The socket was open, but turning the timeout back didn't work.
			// The only possible reason was that the connection was just closed on this side.
			return StatusNotOpen
		}

		return StatusOpen
	}

	return StatusNotOpen // recvErr is not nil or n == 0 with recvErr == nil which means EOF
}
