package device

import (
	"context"
	"io"
	"text/template"

	"github.com/frizz925/wireguard-controller/internal/wireguard"
)

const (
	DEFAULT_ADDRESS = "10.5.0.1"
	DEFAULT_NETWORK = "10.5.0.0"
	DEFAULT_NETMASK = 16
)

type Device interface {
	GetName() string
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

func applyDefaultDevice(dev *device) {
	if dev.Address == "" {
		dev.Address = DEFAULT_ADDRESS
	}
	if dev.Network == "" {
		dev.Network = DEFAULT_NETWORK
	}
	if dev.Netmask == 0 {
		dev.Netmask = DEFAULT_NETMASK
	}
}

func applyRawDevice(dev *device, cfg *Config) {
	dev.Name = cfg.Name
	dev.Address = cfg.Address
	dev.Network = cfg.Network
	dev.Netmask = cfg.Netmask
	dev.PrivateKey = cfg.PrivateKey
	dev.PublicKey = cfg.PublicKey
	dev.tmpl = cfg.Template
	applyDefaultDevice(dev)
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
