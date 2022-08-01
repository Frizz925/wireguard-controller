package client

import (
	"context"

	"github.com/frizz925/wireguard-controller/internal/data"
)

type Repository interface {
	List(ctx context.Context, host, dev string) ([]string, error)
	All(ctx context.Context, host, dev string) ([]*data.Client, error)
	Find(ctx context.Context, host, dev, name string) (*data.Client, error)
	Save(ctx context.Context, host, dev string, client *data.Client) error
	Delete(ctx context.Context, host, dev, name string) error
}
