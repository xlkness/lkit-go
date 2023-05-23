package cli

import (
	"fmt"
	"os"
	"reflect"
)

type Command struct {
	Desc struct {
		Name         string
		ShortDesc    string
		LongDesc     string
		OneLineUsage string // 例如：Example: new [-A aaa|-B bbb] <app>
	} // 指令的描述信息相关
	HasLastOptionArg bool   // 最后一个参数不作为指令，而是指令的参数，例如./lkit new <app>，<app>就是参数了，而不会解析成指令
	LastOptionArg    string // 最后一个参数不作为指令，而是指令的参数，例如./lkit new <app>，<app>就是参数了，而不会解析成指令
	Flag             interface{}
	ExecFun          func(command *Command) error
	SubCommands      map[string]*Command
}

func NewCommand(name, shortDesc, longDesc, onlineUsage string, hasLastOption bool, flag interface{}, execFun func(*Command) error) *Command {
	cmd := new(Command)
	cmd.Desc.Name = name
	cmd.Desc.ShortDesc = shortDesc
	cmd.Desc.LongDesc = longDesc
	cmd.Desc.OneLineUsage = onlineUsage
	cmd.HasLastOptionArg = hasLastOption
	cmd.Flag = flag
	cmd.ExecFun = execFun
	return cmd
}

func (cmd *Command) AddSubCommand(subCmd *Command) *Command {
	if cmd.SubCommands == nil {
		cmd.SubCommands = make(map[string]*Command)
	}
	cmd.SubCommands[subCmd.Desc.Name] = subCmd
	return cmd
}

func (cmd *Command) Execute() error {
	subCommand, args, err := cmd.findSubCommand()
	if err != nil {
		return err
	}
	if len(os.Args) >= 3 {
		if os.Args[len(os.Args)-1] == "-help" ||
			os.Args[len(os.Args)-1] == "-h" {
			fmt.Printf("%v", subCommand.Usage(""))
			os.Exit(0)
		}
	}

	err = extractArgPairs2Flag(subCommand.Flag, args)
	if err != nil {
		return err
	}

	if subCommand.ExecFun != nil {
		return subCommand.ExecFun(subCommand)
	} else {
		fmt.Printf("%s", subCommand.Usage(""))
		return nil
	}
}

func (cmd *Command) findSubCommand() (*Command, []*argPair, error) {
	cmds, args := extractArgs(os.Args[1:])
	var finalCmd *Command = cmd
	for i, v := range cmds {
		if finalCmd.SubCommands != nil {
			tmpCmd, find := finalCmd.SubCommands[v]
			if find {
				finalCmd = tmpCmd
				if len(cmds) >= 2 && i == len(cmds)-2 {
					if finalCmd.HasLastOptionArg {
						finalCmd.LastOptionArg = cmds[i+1]
						return finalCmd, args, nil
					}
				}
				continue
			}
		}
		return nil, nil, fmt.Errorf("nonsupport sub command:%v", v)
	}

	return finalCmd, args, nil
}

func (cmd *Command) Usage(err string) string {
	str := ""
	if err != "" {
		str += fmt.Sprintf("Error: %s\n", err)
	}
	str += fmt.Sprintf("Name: \n")
	str += fmt.Sprintf("  %s - %s\n\n", cmd.Desc.Name, cmd.Desc.ShortDesc)
	str += fmt.Sprintf("  %s\n", cmd.Desc.LongDesc)
	str += fmt.Sprintf("\n")

	str += fmt.Sprintf("Usage: \n")
	str += fmt.Sprintf("  %s\n", cmd.Desc.OneLineUsage)
	str += fmt.Sprintf("\n")

	if len(cmd.SubCommands) != 0 {
		str += fmt.Sprintf("Command: \n")
		for _, v := range cmd.SubCommands {
			str += fmt.Sprintf("  %s\n", rpad(v.Desc.Name+" - "+v.Desc.ShortDesc, 50))
		}

		str += fmt.Sprintf("\n")
	}

	if cmd.Flag != nil {
		flagTo := reflect.TypeOf(cmd.Flag).Elem()
		if flagTo.NumField() > 0 {
			str += fmt.Sprintf("Flags:\n")
			for i := 0; i < flagTo.NumField(); i++ {
				field := flagTo.Field(i)

				str += fmt.Sprintf("  -%s\t%s\n",
					rpad(field.Tag.Get("name")+"  (default:"+field.Tag.Get("default")+")", 30), field.Tag.Get("desc"))
			}
		}

	}

	return str
}

func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}
