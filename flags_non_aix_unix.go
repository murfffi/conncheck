//go:build !windows && !aix

package conncheck

import "syscall"

const flags int = syscall.MSG_PEEK | syscall.MSG_DONTWAIT
