package cli

type argPair struct {
	flag  string
	value string
}

// extractArgs 提取启动参数字符串，将指令与参数分离开
// 例如：new -f 123 app -b true -c sdfds
// 输出：new app -f 123 -b true -c sdfds
func extractArgs(args []string) ([]string, []*argPair) {
	i := 0
	curPair := argPair{}
	subCmd := make([]string, 0)
	argsPair := make([]*argPair, 0)
	for i < len(args) {
		cur := args[i]
		if curPair.flag != "" {
			curPair.value = cur
			argsPair = append(argsPair, &argPair{curPair.flag, curPair.value})
			curPair.flag = ""
			curPair.value = ""
		} else if cur[0] == '-' {
			// 参数，下一个是
			curPair.flag = cur[1:]
		} else {
			// 指令
			subCmd = append(subCmd, cur)
		}
		i++
	}

	return subCmd, argsPair
}
