//go:build linux
// +build linux

package logsyscall

import "syscall"

const (
	LOCK_EX int = 2
	LOCK_UN int = 8
)

// Flock
func Flock(fd int, how int) (err error) {
	return syscall.Flock(fd, how)
}
