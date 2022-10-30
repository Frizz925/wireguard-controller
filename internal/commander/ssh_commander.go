package commander

import (
	"bytes"
	"fmt"

	"github.com/melbahja/goph"
)

type SSHCommander struct {
	client *goph.Client
}

func NewSSHCommander(client *goph.Client) *SSHCommander {
	return &SSHCommander{client}
}

func (sc *SSHCommander) Command(cmd *Command) error {
	var buf bytes.Buffer
	sshCmd, err := sc.client.CommandContext(cmd.Context, cmd.Name, cmd.Args...)
	if err != nil {
		return err
	}
	if cmd.Stdin != nil {
		sshCmd.Stdin = cmd.Stdin
	}
	if cmd.Stdout != nil {
		sshCmd.Stdout = cmd.Stdout
	}
	sshCmd.Stderr = &buf
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err.Error(), buf.String())
	}
	if cmd.Stderr != nil {
		_, err := buf.WriteTo(cmd.Stderr)
		return err
	}
	return nil
}
