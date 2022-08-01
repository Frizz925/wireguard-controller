package command

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type command struct {
	Name   string
	Args   []string
	Input  []byte
	Output []byte
}

func runCommand(ctx context.Context, cmd *command) (int, error) {
	var buf bytes.Buffer
	ec := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	if cmd.Input != nil {
		ec.Stdin = bytes.NewReader(cmd.Input)
	}
	if cmd.Output != nil {
		ec.Stdout = &buf
	}
	if err := ec.Run(); err != nil {
		return 0, err
	}
	if cmd.Output == nil {
		return 0, nil
	}
	return buf.Read(cmd.Output)
}

func processOutput(output []byte) string {
	return strings.TrimSpace(string(output))
}

func OutputCommand(ctx context.Context, output []byte, name string, arg ...string) (int, error) {
	return runCommand(ctx, &command{
		Name:   name,
		Args:   arg,
		Output: output,
	})
}

func OutputStringCommand(ctx context.Context, name string, arg ...string) (string, error) {
	output := make([]byte, 256)
	n, err := OutputCommand(ctx, output, name, arg...)
	if err != nil {
		return "", err
	}
	return processOutput(output[:n]), nil
}

func InputOutputCommand(ctx context.Context, input []byte, output []byte, name string, arg ...string) (int, error) {
	return runCommand(ctx, &command{
		Name:   name,
		Args:   arg,
		Input:  input,
		Output: output,
	})
}

func InputOutputStringCommand(ctx context.Context, input string, name string, arg ...string) (string, error) {
	output := make([]byte, 256)
	n, err := InputOutputCommand(ctx, []byte(input), output, name, arg...)
	if err != nil {
		return "", err
	}
	return processOutput(output[:n]), nil
}
