//go:build !windows && !unix

package conncheck

import "syscall"

func tryPeek(rawConn syscall.RawConn) Status {
	return StatusUnknown
}
