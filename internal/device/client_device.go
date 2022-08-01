package device

import (
	"context"
	"io"

	"github.com/frizz925/wireguard-controller/internal/data"
	clientRepo "github.com/frizz925/wireguard-controller/internal/repositories/client"
)

type clientDevice struct {
	device
	Server       *ServerDevice
	PresharedKey string
	repo         clientRepo.Repository
}

type clientConfig struct {
	Config
	Server       *ServerDevice
	PresharedKey string
	Repository   clientRepo.Repository
}

func newRawClientDevice(cfg *clientConfig) *clientDevice {
	cd := &clientDevice{}
	applyRawDevice(&cd.device, &cfg.Config)
	cd.Server = cfg.Server
	cd.PresharedKey = cfg.PresharedKey
	cd.repo = cfg.Repository
	return cd
}

func newClientDevice(ctx context.Context, cfg *clientConfig) (*clientDevice, error) {
	cd := newRawClientDevice(cfg)
	if err := applyDevice(ctx, &cd.device, &cfg.Config); err != nil {
		return nil, err
	}
	return cd, nil
}

func (sd *clientDevice) GetName() string {
	return sd.Name
}

func (cd *clientDevice) WriteConfig(w io.Writer) error {
	return cd.tmpl.ExecuteTemplate(w, "client", cd)
}

func (cd *clientDevice) Load(ctx context.Context) error {
	data, err := cd.repo.Find(ctx, cd.Server.Host, cd.Server.Name, cd.Name)
	if err != nil {
		return err
	}
	cd.Name = data.Name
	cd.Address = data.Address
	cd.PrivateKey = data.PrivateKey
	cd.PublicKey = data.PublicKey
	cd.PresharedKey = data.PresharedKey
	return nil
}

func (cd *clientDevice) Save(ctx context.Context) error {
	return cd.repo.Save(ctx, cd.Server.Host, cd.Server.Name, &data.Client{
		Name:         cd.Name,
		Address:      cd.Address,
		PrivateKey:   cd.PrivateKey,
		PublicKey:    cd.PublicKey,
		PresharedKey: cd.PresharedKey,
	})
}

func (cd *clientDevice) Delete(ctx context.Context) error {
	return cd.repo.Delete(ctx, cd.Server.Host, cd.Server.Name, cd.Name)
}
