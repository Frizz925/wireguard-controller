package commander

import (
	"context"
	"io"
)

type Command struct {
	Context context.Context
	Name    string
	Args    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

type Commander interface {
	Command(cmd *Command) error
}
