package device

import (
	"context"
	"text/template"

	"github.com/frizz925/wireguard-controller/internal/wireguard"
)

const DEFAULT_ADDRESS = "192.168.128.1"

type device struct {
	Name    string
	Address string

	PrivateKey string
	PublicKey  string

	ctrl wireguard.Controller
	tmpl *template.Template
}

type Config struct {
	Name    string
	Address string

	PrivateKey string
	PublicKey  string

	Controller wireguard.Controller
	Template   *template.Template
}

func applyDefaultDevice(dev *device) {
	if dev.Address == "" {
		dev.Address = DEFAULT_ADDRESS
	}
}

func applyRawDevice(dev *device, cfg *Config) {
	dev.Name = cfg.Name
	dev.Address = cfg.Address
	dev.PrivateKey = cfg.PrivateKey
	dev.PublicKey = cfg.PublicKey
	dev.ctrl = cfg.Controller
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
	d.PrivateKey, err = d.ctrl.Genkey(ctx)
	if err != nil {
		return err
	}
	d.PublicKey, err = d.ctrl.Pubkey(ctx, d.PrivateKey)
	return err
}
