package cli

import (
	"fmt"
	"testing"
)

func TestFlag(t *testing.T) {
	type Flag struct {
		A string `name:"a" desc"xxx" default:"aaa"`
		B int    `name:"b" desc:"xxx" default:"123"`
	}
	flag := &Flag{}
	extractArgPairs2Flag(flag, []*argPair{
		{"a", "aaaaa"},
		{"b", "234"},
	})

	fmt.Printf("%+v\n", flag)
}
