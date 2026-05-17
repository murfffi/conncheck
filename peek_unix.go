//go:build !windows

package conncheck

import (
	"errors"
	"io"
	"syscall"
)

func tryPeek(rawConn syscall.RawConn) Status {
	readErr := tryPeekUnix(rawConn)
	if readErr != nil {
		return StatusNotOpen
	}
	return StatusOpen
}

func tryPeekUnix(rawConn syscall.RawConn) (err error) {
	var n int

	if readErr := rawConn.Read(func(fd uintptr) bool {
		var buffer [1]byte
		n, _, err = syscall.Recvfrom(int(fd), buffer[:], flags)
		return true
	}); readErr != nil {
		return readErr
	}

	if n > 0 {
		return nil // connection is open and there is something in the buffer
	}

	if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
		// connection is open and there is nothing in the buffer
		return nil
	}

	if err != nil {
		return err
	}

	return io.EOF // n = 0
}
