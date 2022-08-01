package device

import (
	"context"
	"io"
	"net"

	"github.com/frizz925/wireguard-controller/internal/data"
	clientRepo "github.com/frizz925/wireguard-controller/internal/repositories/client"
	serverRepo "github.com/frizz925/wireguard-controller/internal/repositories/server"
)

type ServerDevice struct {
	device

	Host       string
	ListenPort int

	serverRepo serverRepo.Repository
	clientRepo clientRepo.Repository

	clients map[string]*clientDevice
}

type ServerConfig struct {
	Config

	Host       string
	ListenPort int

	ServerRepo serverRepo.Repository
	ClientRepo clientRepo.Repository
}

func NewRawServerDevice(cfg *ServerConfig) *ServerDevice {
	sd := &ServerDevice{}
	applyRawDevice(&sd.device, &cfg.Config)
	sd.Host = cfg.Host
	sd.ListenPort = cfg.ListenPort
	sd.serverRepo = cfg.ServerRepo
	sd.clientRepo = cfg.ClientRepo
	sd.clients = make(map[string]*clientDevice)
	return sd
}

func NewServerDevice(ctx context.Context, cfg *ServerConfig) (*ServerDevice, error) {
	sd := NewRawServerDevice(cfg)
	if err := applyDevice(ctx, &sd.device, &cfg.Config); err != nil {
		return nil, err
	}
	return sd, nil
}

func (sd *ServerDevice) AddClient(ctx context.Context, name string) (Device, error) {
	octet := len(sd.clients) + 2
	ip := net.ParseIP(sd.Network)
	if ip == nil {
		return nil, &net.ParseError{}
	}

	ip = ip.To4()
	if octet > 255 {
		temp := octet / 256
		ip[2] = byte(temp)
		octet -= temp * 256
	}
	ip[3] = byte(octet)

	cd, err := newClientDevice(ctx, &clientConfig{
		Config: Config{
			Name:     name,
			Network:  sd.Network,
			Address:  ip.String(),
			Netmask:  sd.Netmask,
			Template: sd.tmpl,
		},
		Server:     sd,
		Repository: sd.clientRepo,
	})
	if err != nil {
		return nil, err
	}
	if err := cd.generatePresharedKey(ctx); err != nil {
		return nil, err
	}

	sd.clients[name] = cd
	return cd, nil
}

func (sd *ServerDevice) GetClient(name string) Device {
	return sd.clients[name]
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
