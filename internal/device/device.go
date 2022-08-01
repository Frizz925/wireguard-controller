package device

import (
	"context"
	"io"
	"text/template"

	"github.com/frizz925/wireguard-controller/internal/wireguard"
)

type Device interface {
	WriteConfig(w io.Writer) error
}

type device struct {
	Name    string
	Address string
	Network string
	Netmask int

	PrivateKey string
	PublicKey  string

	tmpl *template.Template
}

type Config struct {
	Name    string
	Address string
	Network string
	Netmask int

	PrivateKey string
	PublicKey  string

	Template *template.Template
}

func applyRawDevice(dev *device, cfg *Config) {
	dev.Name = cfg.Name
	dev.Address = cfg.Address
	dev.Network = cfg.Network
	dev.Netmask = cfg.Netmask
	dev.PrivateKey = cfg.PrivateKey
	dev.PublicKey = cfg.PublicKey
	dev.tmpl = cfg.Template
}

func applyDevice(ctx context.Context, dev *device, cfg *Config) error {
	applyRawDevice(dev, cfg)
	if dev.PrivateKey != "" {
		return nil
	}
	return dev.generateKeys(ctx)
}

func (d *device) generateKeys(ctx context.Context) error {
	var err error
	d.PrivateKey, err = wireguard.Genkey(ctx)
	if err != nil {
		return err
	}
	d.PublicKey, err = wireguard.Pubkey(ctx, d.PrivateKey)
	return err
}
