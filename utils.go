package lkit_go

import "github.com/xlkness/lkit-go/internal/utils"

func StringLowerCase(s string) string {
	return utils.LowerCase(s)
}

func StringCamelCase(s string) string {
	return utils.CamelCase(s)
}
