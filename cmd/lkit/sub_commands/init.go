package sub_commands

import (
	new2 "github.com/xlkness/lkit-go/cmd/lkit/sub_commands/new"
	"github.com/xlkness/lkit-go/internal/cli"
)

func SubCommands() []*cli.Command {
	subCommands := []*cli.Command{
		new2.CommandNew(),
	}
	return subCommands
}
