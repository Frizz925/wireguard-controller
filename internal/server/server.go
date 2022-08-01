package server

import (
	"context"
	"errors"
	"path"
	"text/template"

	"github.com/frizz925/wireguard-controller/internal/device"
	clientRepo "github.com/frizz925/wireguard-controller/internal/repositories/client"
	serverRepo "github.com/frizz925/wireguard-controller/internal/repositories/server"
)

var ErrNotFound = errors.New("not found")

type Server struct {
	Host string

	tmpl *template.Template

	serverRepo serverRepo.Repository
	clientRepo clientRepo.Repository
	devices    map[string]*device.ServerDevice
}

type Config struct {
	Host         string
	TemplatesDir string
	ServerRepo   serverRepo.Repository
	ClientRepo   clientRepo.Repository
}

func New(cfg *Config) (*Server, error) {
	tmpl, err := template.ParseGlob(path.Join(cfg.TemplatesDir, "*.tmpl"))
	if err != nil {
		return nil, err
	}
	return &Server{
		Host:       cfg.Host,
		tmpl:       tmpl,
		serverRepo: cfg.ServerRepo,
		clientRepo: cfg.ClientRepo,
		devices:    make(map[string]*device.ServerDevice),
	}, nil
}

func (s *Server) AddDevice(ctx context.Context, name string, port int) (*device.ServerDevice, error) {
	sd, err := device.NewServerDevice(ctx, &device.ServerConfig{
		Config: device.Config{
			Name:     name,
			Template: s.tmpl,
		},
		Host:       s.Host,
		ListenPort: port,
		ServerRepo: s.serverRepo,
		ClientRepo: s.clientRepo,
	})
	if err != nil {
		return nil, err
	}
	s.devices[name] = sd
	return sd, nil
}

func (s *Server) GetDevice(name string) *device.ServerDevice {
	return s.devices[name]
}

func (s *Server) Load(ctx context.Context) error {
	names, err := s.serverRepo.List(ctx, s.Host)
	if err != nil {
		return err
	}
	for _, name := range names {
		sd := device.NewRawServerDevice(&device.ServerConfig{
			Config: device.Config{
				Name:     name,
				Template: s.tmpl,
			},
			Host:       s.Host,
			ServerRepo: s.serverRepo,
			ClientRepo: s.clientRepo,
		})
		if err := sd.Load(ctx); err != nil {
			return err
		}
		s.devices[name] = sd
	}
	return nil
}

func (s *Server) Save(ctx context.Context) error {
	for _, dev := range s.devices {
		if err := dev.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
