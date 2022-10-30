package commander

import (
	"os/exec"
)

type LocalCommander struct{}

func NewLocalCommander() *LocalCommander {
	return &LocalCommander{}
}

func (LocalCommander) Command(cmd *Command) error {
	ec := exec.CommandContext(cmd.Context, cmd.Name, cmd.Args...)
	if cmd.Stdin != nil {
		ec.Stdin = cmd.Stdin
	}
	if cmd.Stdout != nil {
		ec.Stdout = cmd.Stdout
	}
	if cmd.Stderr != nil {
		ec.Stderr = cmd.Stderr
	}
	return ec.Run()
}
