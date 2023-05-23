package main

import (
	"github.com/xlkness/lkit-go/cmd/lkit/sub_commands"
	"github.com/xlkness/lkit-go/internal/cli"
)

var rootCmd = cli.NewCommand("lkit", "lkit集成命令行工具", "", "./lkit [-h|-help] || ./lkit [command] [-h|-help]", false, nil, nil)

func init() {
	for _, subCommand := range sub_commands.SubCommands() {
		rootCmd.AddSubCommand(subCommand)
	}
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
