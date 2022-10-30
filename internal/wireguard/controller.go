package wireguard

import "context"

type Controller interface {
	Genkey(ctx context.Context) (string, error)
	Pubkey(ctx context.Context, privkey string) (string, error)
	Genpsk(ctx context.Context) (string, error)
	Device(name string) DeviceController
}

type DeviceController interface {
	SaveConfig(ctx context.Context, content []byte) error
	IsEnabled(ctx context.Context) (bool, error)
	IsActive(ctx context.Context) (bool, error)
	Enable(ctx context.Context) error
	Start(ctx context.Context) error
	Restart(ctx context.Context) error
}
