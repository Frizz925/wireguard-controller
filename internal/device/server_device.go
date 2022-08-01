package device

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/frizz925/wireguard-controller/internal/data"
	clientRepo "github.com/frizz925/wireguard-controller/internal/repositories/client"
	serverRepo "github.com/frizz925/wireguard-controller/internal/repositories/server"
	"github.com/frizz925/wireguard-controller/internal/wireguard"
)

const DEFAULT_LISTEN_PORT = 51820

var ErrNotFound = errors.New("not found")

type ServerDevice struct {
	device

	Host       string
	ListenPort int

	serverRepo serverRepo.Repository
	clientRepo clientRepo.Repository

	clients     map[string]*clientDevice
	lastAddress string
}

type ServerConfig struct {
	Config

	Host       string
	ListenPort int

	ServerRepo serverRepo.Repository
	ClientRepo clientRepo.Repository
}

func applyDefaultServerDevice(sd *ServerDevice) {
	if sd.ListenPort == 0 {
		sd.ListenPort = DEFAULT_LISTEN_PORT
	}
}

func NewRawServerDevice(cfg *ServerConfig) *ServerDevice {
	sd := &ServerDevice{}
	applyRawDevice(&sd.device, &cfg.Config)
	sd.Host = cfg.Host
	sd.ListenPort = cfg.ListenPort
	sd.serverRepo = cfg.ServerRepo
	sd.clientRepo = cfg.ClientRepo
	sd.clients = make(map[string]*clientDevice)
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

func (sd *ServerDevice) GetName() string {
	return sd.Name
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

func (sd *ServerDevice) GetClient(name string) Device {
	v, ok := sd.clients[name]
	if ok {
		return v
	}
	return nil
}

func (sd *ServerDevice) HasClient(name string) bool {
	return sd.GetClient(name) != nil
}

func (sd *ServerDevice) AddClient(ctx context.Context, name string) (Device, error) {
	lip := sd.lastAddress
	if lip == "" {
		lip = sd.Address
	}

	ip := net.ParseIP(lip).To4()
	octet := ip[3]
	if octet >= 255 {
		ip[2]++
		ip[3] = 0
	} else {
		ip[3]++
	}

	psk, err := wireguard.Genpsk(ctx)
	if err != nil {
		return nil, err
	}
	cd, err := newClientDevice(ctx, &clientConfig{
		Config: Config{
			Name:     name,
			Network:  sd.Network,
			Address:  ip.String(),
			Netmask:  sd.Netmask,
			Template: sd.tmpl,
		},
		Server:       sd,
		PresharedKey: psk,
		Repository:   sd.clientRepo,
	})
	if err != nil {
		return nil, err
	}

	sd.lastAddress = ip.String()
	sd.clients[name] = cd
	return cd, nil
}

func (sd *ServerDevice) RemoveClient(ctx context.Context, name string) (Device, error) {
	cd, ok := sd.clients[name]
	if !ok {
		return nil, ErrNotFound
	}
	if err := cd.Delete(ctx); err != nil {
		return nil, err
	}
	delete(sd.clients, name)
	return cd, nil
}

func (sd *ServerDevice) writePeerConfig(w io.Writer, peer *clientDevice) error {
	return sd.tmpl.ExecuteTemplate(w, "server_peer", peer)
}

func (sd *ServerDevice) Load(ctx context.Context) error {
	data, err := sd.serverRepo.Find(ctx, sd.Host, sd.Name)
	if err != nil {
		return err
	}
	sd.Host = data.Host
	sd.Name = data.Name
	sd.ListenPort = data.ListenPort
	sd.Address = data.Address
	sd.Network = data.Network
	sd.Netmask = data.Netmask
	sd.PrivateKey = data.PrivateKey
	sd.PublicKey = data.PublicKey
	sd.lastAddress = data.LastAddress

	names, err := sd.clientRepo.List(ctx, sd.Host, sd.Name)
	if err != nil {
		return err
	}
	for _, name := range names {
		cd := newRawClientDevice(&clientConfig{
			Config: Config{
				Name:     name,
				Network:  sd.Network,
				Netmask:  sd.Netmask,
				Template: sd.tmpl,
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
	err := sd.serverRepo.Save(ctx, &data.Server{
		Host:       sd.Host,
		Name:       sd.Name,
		ListenPort: sd.ListenPort,

		Address: sd.Address,
		Network: sd.Network,
		Netmask: sd.Netmask,

		PrivateKey: sd.PrivateKey,
		PublicKey:  sd.PublicKey,

		LastAddress: sd.lastAddress,
	})
	if err != nil {
		return err
	}
	for _, client := range sd.clients {
		if err := client.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
