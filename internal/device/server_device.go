package device

import (
	"context"
	"errors"
	"io"

	"github.com/frizz925/wireguard-controller/internal/config"
	"github.com/frizz925/wireguard-controller/internal/data"
	clientRepo "github.com/frizz925/wireguard-controller/internal/repositories/client"
	serverRepo "github.com/frizz925/wireguard-controller/internal/repositories/server"
)

const (
	DEFAULT_NETWORK     = "192.168.128.0"
	DEFAULT_NETMASK     = 24
	DEFAULT_LISTEN_PORT = 51820
)

var ErrNotFound = errors.New("not found")

type ServerDevice struct {
	device

	Host       string
	Network    string
	Netmask    int
	DNS        string
	ListenPort int

	PostUp  string
	PreDown string

	serverRepo serverRepo.Repository
	clientRepo clientRepo.Repository

	clients map[string]*ClientDevice
}

type ServerConfig struct {
	Config

	Host       string
	Network    string
	Netmask    int
	DNS        string
	ListenPort int

	PostUp  string
	PreDown string

	ServerRepo serverRepo.Repository
	ClientRepo clientRepo.Repository
}

func applyDefaultServerDevice(sd *ServerDevice) {
	if sd.Network == "" {
		sd.Network = DEFAULT_NETWORK
	}
	if sd.Netmask == 0 {
		sd.Netmask = DEFAULT_NETMASK
	}
	if sd.ListenPort <= 0 {
		sd.ListenPort = DEFAULT_LISTEN_PORT
	}
	if sd.DNS == "" {
		sd.DNS = sd.Address
	}
}

func NewRawServerDevice(cfg *ServerConfig) *ServerDevice {
	sd := &ServerDevice{}
	applyRawDevice(&sd.device, &cfg.Config)
	sd.Host = cfg.Host
	sd.Network = cfg.Network
	sd.Netmask = cfg.Netmask
	sd.DNS = cfg.DNS
	sd.ListenPort = cfg.ListenPort
	sd.PostUp = cfg.PostUp
	sd.PreDown = cfg.PreDown
	sd.serverRepo = cfg.ServerRepo
	sd.clientRepo = cfg.ClientRepo
	sd.clients = make(map[string]*ClientDevice)
	applyDefaultServerDevice(sd)
	return sd
}

func NewServerDevice(ctx context.Context, cfg *ServerConfig) (*ServerDevice, error) {
	sd := NewRawServerDevice(cfg)
	if err := applyDevice(ctx, &sd.device, &cfg.Config); err != nil {
		return nil, err
	}
	return sd, nil
}

func (sd *ServerDevice) Apply(cfg config.Device) {
	sd.Address = cfg.Address
	sd.Network = cfg.Network
	sd.Netmask = cfg.Netmask
	sd.PostUp = cfg.PostUp
	sd.PreDown = cfg.PreDown
	if cfg.ListenPort > 0 {
		sd.ListenPort = cfg.ListenPort
	}
	if cfg.DNS != "" {
		sd.DNS = cfg.DNS
	} else {
		sd.DNS = cfg.Address
	}
}

func (sd *ServerDevice) WriteConfig(w io.Writer) error {
	if err := sd.tmpl.ExecuteTemplate(w, "server_head", sd); err != nil {
		return err
	}
	for _, client := range sd.clients {
		if err := sd.writePeerConfig(w, client); err != nil {
			return err
		}
	}
	return nil
}

func (sd *ServerDevice) GetClientNames() []string {
	names := make([]string, 0, len(sd.clients))
	for name := range sd.clients {
		names = append(names, name)
	}
	return names
}

func (sd *ServerDevice) GetClient(name string) *ClientDevice {
	v, ok := sd.clients[name]
	if ok {
		return v
	}
	return nil
}

func (sd *ServerDevice) HasClient(name string) bool {
	return sd.GetClient(name) != nil
}

func (sd *ServerDevice) AddClient(ctx context.Context, user config.User) (*ClientDevice, error) {
	psk, err := sd.device.ctrl.Genpsk(ctx)
	if err != nil {
		return nil, err
	}
	cd, err := NewClientDevice(ctx, &clientConfig{
		Config: Config{
			Name:       user.Name,
			Address:    user.Address,
			Controller: sd.ctrl,
			Template:   sd.tmpl,
		},
		Server:       sd,
		PresharedKey: psk,
		Repository:   sd.clientRepo,
	})
	if err != nil {
		return nil, err
	}
	if err := cd.Save(ctx); err != nil {
		return nil, err
	}
	sd.clients[user.Name] = cd
	return cd, nil
}

func (sd *ServerDevice) RemoveClient(ctx context.Context, name string) (*ClientDevice, error) {
	cd, ok := sd.clients[name]
	if !ok {
		return nil, ErrNotFound
	}
	delete(sd.clients, name)
	if err := cd.Delete(ctx); err != nil {
		return nil, err
	}
	return cd, nil
}

func (sd *ServerDevice) writePeerConfig(w io.Writer, peer *ClientDevice) error {
	return sd.tmpl.ExecuteTemplate(w, "server_peer", peer)
}

func (sd *ServerDevice) Load(ctx context.Context) error {
	data, err := sd.serverRepo.Find(ctx, sd.Host, sd.Name)
	if err != nil {
		return err
	}
	sd.PrivateKey = data.PrivateKey
	sd.PublicKey = data.PublicKey

	names, err := sd.clientRepo.List(ctx, sd.Host, sd.Name)
	if err != nil {
		return err
	}
	for _, name := range names {
		cd := NewRawClientDevice(&clientConfig{
			Config: Config{
				Name:       name,
				Controller: sd.ctrl,
				Template:   sd.tmpl,
			},
			Server:     sd,
			Repository: sd.clientRepo,
		})
		if err := cd.Load(ctx); err != nil {
			return err
		}
		sd.clients[name] = cd
	}
	return nil
}

func (sd *ServerDevice) Save(ctx context.Context) error {
	return sd.serverRepo.Save(ctx, sd.Host, sd.Name, &data.Server{
		PrivateKey: sd.PrivateKey,
		PublicKey:  sd.PublicKey,
	})
}
