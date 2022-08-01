package client

import (
	"context"
	"path"

	"github.com/frizz925/wireguard-controller/internal/data"
	"github.com/frizz925/wireguard-controller/internal/storage"
)

type LocalRepository struct {
	storage *storage.LocalStorage
}

func NewLocalRepository(dir string) *LocalRepository {
	return &LocalRepository{
		storage: storage.NewLocalStorage(dir),
	}
}

func (r *LocalRepository) List(ctx context.Context, host, dev string) ([]string, error) {
	return r.storage.List(ctx, r.getPrefix(host, dev))
}

func (r *LocalRepository) All(ctx context.Context, host, dev string) ([]*data.Client, error) {
	names, err := r.List(ctx, host, dev)
	if err != nil {
		return nil, err
	}
	results := make([]*data.Client, len(names))
	for idx, name := range names {
		v, err := r.Find(ctx, host, dev, name)
		if err != nil {
			return nil, err
		}
		results[idx] = v
	}
	return results, nil
}

func (r *LocalRepository) Find(ctx context.Context, host, dev, name string) (*data.Client, error) {
	data := &data.Client{}
	err := r.storage.Load(ctx, r.getPath(host, dev, name), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *LocalRepository) Save(ctx context.Context, host, dev string, client *data.Client) error {
	return r.storage.Save(ctx, r.getPath(host, dev, client.Name), client)
}

func (r *LocalRepository) getPrefix(host, dev string) string {
	return path.Join(host, dev)
}

func (r *LocalRepository) getPath(host, dev, name string) string {
	return path.Join(r.getPrefix(host, dev), name)
}
