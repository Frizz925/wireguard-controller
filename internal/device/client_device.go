package device

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/frizz925/wireguard-controller/internal/config"
	"github.com/frizz925/wireguard-controller/internal/data"
	clientRepo "github.com/frizz925/wireguard-controller/internal/repositories/client"
)

type ClientDevice struct {
	device
	Server *ServerDevice

	PresharedKey string
	AllowedIPs   string

	repo clientRepo.Repository
}

type clientConfig struct {
	Config
	Server *ServerDevice

	PresharedKey string
	AllowedIPs   string

	Repository clientRepo.Repository
}

func applyDefaultClientDevice(cd *ClientDevice) {
	if cd.AllowedIPs == "" {
		cd.AllowedIPs = cd.defaultAllowedIPs()
	}
}

func NewRawClientDevice(cfg *clientConfig) *ClientDevice {
	cd := &ClientDevice{}
	applyRawDevice(&cd.device, &cfg.Config)
	cd.Server = cfg.Server
	cd.PresharedKey = cfg.PresharedKey
	cd.AllowedIPs = cfg.AllowedIPs
	cd.repo = cfg.Repository
	applyDefaultClientDevice(cd)
	return cd
}

func NewClientDevice(ctx context.Context, cfg *clientConfig) (*ClientDevice, error) {
	cd := NewRawClientDevice(cfg)
	if err := applyDevice(ctx, &cd.device, &cfg.Config); err != nil {
		return nil, err
	}
	return cd, nil
}

func (cd *ClientDevice) WriteConfig(w io.Writer) error {
	return cd.tmpl.ExecuteTemplate(w, "client", cd)
}

func (cd *ClientDevice) Load(ctx context.Context) error {
	data, err := cd.repo.Find(ctx, cd.Server.Host, cd.Server.Name, cd.Name)
	if err != nil {
		return err
	}
	cd.PrivateKey = data.PrivateKey
	cd.PublicKey = data.PublicKey
	cd.PresharedKey = data.PresharedKey
	return nil
}

func (cd *ClientDevice) Save(ctx context.Context) error {
	return cd.repo.Save(ctx, cd.Server.Host, cd.Server.Name, cd.Name, &data.Client{
		PrivateKey:   cd.PrivateKey,
		PublicKey:    cd.PublicKey,
		PresharedKey: cd.PresharedKey,
	})
}

func (cd *ClientDevice) Delete(ctx context.Context) error {
	return cd.repo.Delete(ctx, cd.Server.Host, cd.Server.Name, cd.Name)
}

func (cd *ClientDevice) Apply(cfg config.User) {
	cd.Name = cfg.Name
	cd.Address = cfg.Address
	if len(cfg.AllowedIPs) > 0 {
		cd.AllowedIPs = strings.Join(cfg.AllowedIPs, ",")
	} else {
		cd.AllowedIPs = cd.defaultAllowedIPs()
	}
}

func (cd *ClientDevice) defaultAllowedIPs() string {
	return fmt.Sprintf("%s/32", cd.Address)
}
