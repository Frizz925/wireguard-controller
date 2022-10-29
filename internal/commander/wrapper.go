package commander

import (
	"bytes"
	"context"
	"io"
	"strings"
)

type Wrapper struct {
	Commander
}

func NewWrapper(cmd Commander) *Wrapper {
	return &Wrapper{cmd}
}

func (w *Wrapper) SimpleCommand(ctx context.Context, name string, args ...string) error {
	return w.Command(&Command{
		Context: ctx,
		Name:    name,
		Args:    args,
	})
}

func (w *Wrapper) OutputCommand(ctx context.Context, output io.Writer, name string, args ...string) error {
	return w.Command(&Command{
		Context: ctx,
		Name:    name,
		Args:    args,
		Stdout:  output,
	})
}

func (w *Wrapper) OutputStringCommand(ctx context.Context, name string, args ...string) (string, error) {
	return w.stringCommand(ctx, "", name, args)
}

func (w *Wrapper) InputCommand(ctx context.Context, input io.Reader, name string, args ...string) error {
	return w.Command(&Command{
		Context: ctx,
		Name:    name,
		Args:    args,
		Stdin:   input,
	})
}

func (w *Wrapper) InputStringCommand(ctx context.Context, input string, name string, args ...string) error {
	_, err := w.stringCommand(ctx, input, name, args)
	return err
}

func (w *Wrapper) InputOutputCommand(ctx context.Context, input io.Reader, output io.Writer, name string, args ...string) error {
	return w.Command(&Command{
		Context: ctx,
		Name:    name,
		Args:    args,
		Stdin:   input,
		Stdout:  output,
	})
}

func (w *Wrapper) InputOutputStringCommand(ctx context.Context, input string, name string, args ...string) (string, error) {
	return w.stringCommand(ctx, input, name, args)
}

func (w *Wrapper) stringCommand(ctx context.Context, input string, name string, args []string) (string, error) {
	var buf bytes.Buffer
	cmd := &Command{
		Context: ctx,
		Name:    name,
		Args:    args,
		Stdout:  &buf,
	}
	if input != "" {
		cmd.Stdin = bytes.NewReader([]byte(input))
	}
	if err := w.Command(cmd); err != nil {
		return "", nil
	}
	return strings.TrimSpace(buf.String()), nil
}
