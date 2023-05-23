package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestArgsExtract(t *testing.T) {
	args := "new -f 123 app -b true -c sdfds"
	tokens := strings.Split(args, " ")
	fmt.Printf("tokens:%+v\n", tokens)
	a, b := extractArgs(tokens)
	fmt.Printf("%+v\n", a)
	for _, v := range b {
		fmt.Printf("%+v ", v)
	}
	fmt.Println()
}
