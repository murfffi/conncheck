//go:build unix && !aix

package conncheck

import "syscall"

const flags int = syscall.MSG_PEEK | syscall.MSG_DONTWAIT
