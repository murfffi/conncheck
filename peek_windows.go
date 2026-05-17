//go:build windows

package conncheck

import (
	"errors"
	"io"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const MSG_PEEK uint32 = windows.MSG_PEEK

//Use the following code to find the value of FIONBIO
//#include <winsock.h>
//
//int main()
//{
//	printf("FIONBIO is size %u with value %ld (0x%x)\n",
//		sizeof(FIONBIO), FIONBIO, FIONBIO);
//}

const FIONBIO uint32 = 0x8004667e

func tryPeek(rawConn syscall.RawConn) (err error) {
	var n uint32

	if readErr := rawConn.Read(func(fd uintptr) bool {
		h := windows.Handle(fd)

		// TODO: find a way to do this without switching to non-blocking because
		// it is hard to switch back - hard to tell the previous state; maybe ioctl FIONREAD
		err = switchNonBlocking(h)
		if err != nil {
			return true
		}
		buf := [1]byte{}
		wsabuf := windows.WSABuf{
			Len: uint32(len(buf)),
			Buf: &buf[0],
		}

		err = windows.WSARecv(h, &wsabuf, 1, &n, new(MSG_PEEK), nil, nil)
		return true
	}); readErr != nil {
		return readErr
	}

	if n > 0 {
		return nil // connection is open and there is something in the buffer
	}

	// Ignore also windows.WSAEMSGSIZE based on the example in
	// https://github.com/golang/go/blob/364de84f/src/internal/poll/fd_windows.go#L1262

	if errors.Is(err, windows.WSAEWOULDBLOCK) || errors.Is(err, windows.WSAEMSGSIZE) {
		// connection is open and there is nothing in the buffer
		return nil
	}

	return io.EOF // n = 0
}

func switchNonBlocking(h windows.Handle) error {
	nonblocking := uint32(1)
	var r uint32
	err := windows.WSAIoctl(
		h,
		FIONBIO,
		(*byte)(unsafe.Pointer(&nonblocking)),
		uint32(unsafe.Sizeof(nonblocking)),
		nil,
		0,
		&r,
		nil,
		0,
	)
	return err
}
