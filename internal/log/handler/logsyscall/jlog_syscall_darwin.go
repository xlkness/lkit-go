//go:build darwin
// +build darwin

package logsyscall

const (
	LOCK_EX int = 2
	LOCK_UN int = 8
)

// Flock
func Flock(fd int, how int) (err error) {
	return nil
}
