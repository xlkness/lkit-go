package utils

import (
	"fmt"
	"os"
)

func OutputInfo(format string, args ...any) {
	ColorPrint(Color_Yellow, fmt.Sprintf("==> "+format+"\n", args...))
}

func OutputError(format string, args ...any) {
	ColorPrint(Color_Red, fmt.Sprintf("==> "+format+"\n", args...))
	os.Exit(1)
}

var (
	Color_Black  int = 30
	Color_Red    int = 31
	Color_Green  int = 32
	Color_Yellow int = 33
	Color_Blue   int = 34
	Color_White  int = 37
)

// 输出有颜色的字体
func ColorPrint(color int, s string) {
	fmt.Printf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 0, color, s, 0x1B)
}
