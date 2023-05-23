package prom

import (
	"testing"
)

func TestProm(t *testing.T) {
	NewEngine(":9008", true).Run()
}
