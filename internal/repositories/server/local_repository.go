package server

import (
	"context"
	"path"

	"github.com/frizz925/wireguard-controller/internal/data"
	"github.com/frizz925/wireguard-controller/internal/storage"
)

type LocalRepository struct {
	storage *storage.LocalStorage
}

func NewLocalRepository(dirs ...string) *LocalRepository {
	return &LocalRepository{
		storage: storage.NewLocalStorage(dirs...),
	}
}

func (r *LocalRepository) List(ctx context.Context, host string) ([]string, error) {
	return r.storage.List(ctx, host)
}

func (r *LocalRepository) Find(ctx context.Context, host, dev string) (*data.Server, error) {
	data := &data.Server{}
	err := r.storage.Load(ctx, r.getPath(host, dev), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *LocalRepository) Save(ctx context.Context, server *data.Server) error {
	return r.storage.Save(ctx, r.getPath(server.Host, server.Name), server)
}

func (r *LocalRepository) getPath(host, dev string) string {
	return path.Join(host, dev)
}
