//go:build aix

package conncheck

const flags int = syscall.MSG_PEEK | syscall.MSG_NONBLOCK
