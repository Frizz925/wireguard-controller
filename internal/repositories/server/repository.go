package server

import (
	"context"

	"github.com/frizz925/wireguard-controller/internal/data"
)

type Repository interface {
	List(ctx context.Context, host string) ([]string, error)
	Find(ctx context.Context, host, dev string) (*data.Server, error)
	Save(ctx context.Context, server *data.Server) error
}
