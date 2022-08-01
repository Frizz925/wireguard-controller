package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/frizz925/wireguard-controller/internal/data"
	"github.com/frizz925/wireguard-controller/internal/device"
	"github.com/frizz925/wireguard-controller/internal/server"
	"github.com/skip2/go-qrcode"
	"gopkg.in/yaml.v3"

	clientRepoPkg "github.com/frizz925/wireguard-controller/internal/repositories/client"
	serverRepoPkg "github.com/frizz925/wireguard-controller/internal/repositories/server"
)

type serverConfig struct {
	Host string
	Cwd  string
	Dir  string

	ServerRepo serverRepoPkg.Repository
	ClientRepo clientRepoPkg.Repository
}

type deviceConfig struct {
	data.Config
	Host   string
	Name   string
	Dir    string
	Server *server.Server
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	serverRepo := serverRepoPkg.NewLocalRepository()
	clientRepo := clientRepoPkg.NewLocalRepository()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfgDir := path.Join(cwd, "configs")

	dirs, err := os.ReadDir(cfgDir)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		host := dir.Name()
		scfg := &serverConfig{
			Host:       host,
			Cwd:        cwd,
			Dir:        path.Join(cfgDir, host),
			ServerRepo: serverRepo,
			ClientRepo: clientRepo,
		}
		if err := generateServer(ctx, scfg); err != nil {
			return err
		}
	}
	return nil
}

func generateServer(ctx context.Context, cfg *serverConfig) error {
	srv, err := server.New(&server.Config{
		Host:         cfg.Host,
		TemplatesDir: path.Join(cfg.Cwd, "templates"),
		ServerRepo:   cfg.ServerRepo,
		ClientRepo:   cfg.ClientRepo,
	})
	if err != nil {
		return err
	}
	if err := srv.Load(ctx); err != nil {
		return err
	}

	files, err := filepath.Glob(path.Join(cfg.Dir, "*.yaml"))
	if err != nil {
		return err
	}
	for _, filePath := range files {
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		var fcfg data.Config
		if err := yaml.NewDecoder(f).Decode(&fcfg); err != nil {
			return err
		}

		file := path.Base(filePath)
		idx := strings.Index(file, ".yaml")
		if idx <= 0 {
			continue
		}
		name := file[:idx]

		dcfg := &deviceConfig{
			Config: fcfg,
			Host:   cfg.Host,
			Name:   name,
			Dir:    path.Join(cfg.Dir, name),
			Server: srv,
		}
		if err := generateDevice(ctx, dcfg); err != nil {
			return err
		}
	}
	return nil
}

func generateDevice(ctx context.Context, cfg *deviceConfig) error {
	var err error
	srv := cfg.Server

	dev := srv.GetDevice(cfg.Name)
	if dev == nil {
		dev, err = srv.AddDevice(ctx, cfg.Name, cfg.ListenPort)
		if err != nil {
			return err
		}
	}
	dev.Host = cfg.Host
	if cfg.Address != "" {
		dev.Address = cfg.Address
	}
	if cfg.Network != "" {
		dev.Network = cfg.Network
	}
	if cfg.Netmask != 0 {
		dev.Netmask = cfg.Netmask
	}
	if cfg.ListenPort != 0 {
		dev.ListenPort = cfg.ListenPort
	}

	peers := make([]device.Device, len(cfg.Users))
	userMap := make(map[string]device.Device)
	for idx, user := range cfg.Users {
		peer := dev.GetClient(user)
		if peer == nil {
			peer, err = dev.AddClient(ctx, user)
			if err != nil {
				return err
			}
		}
		peers[idx] = peer
		userMap[user] = peer
	}

	// Check for removed user
	for _, user := range dev.GetClientNames() {
		if _, ok := userMap[user]; ok {
			continue
		}
		if _, err := dev.RemoveClient(ctx, user); err != nil {
			return err
		}
	}

	if err := srv.Save(ctx); err != nil {
		return err
	}

	if _, err := os.Stat(cfg.Dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if err := os.RemoveAll(cfg.Dir); err != nil {
			return err
		}
	}
	if err := os.Mkdir(cfg.Dir, 0700); err != nil {
		return err
	}

	var buf bytes.Buffer
	for _, peer := range peers {
		if err := peer.WriteConfig(&buf); err != nil {
			return err
		}

		prefix := path.Join(cfg.Dir, peer.GetName())
		if err := os.WriteFile(fmt.Sprintf("%s.conf", prefix), buf.Bytes(), 0600); err != nil {
			return err
		}
		if err := qrcode.WriteFile(buf.String(), qrcode.Medium, 512, fmt.Sprintf("%s.png", prefix)); err != nil {
			return err
		}
		buf.Reset()
	}

	filename := fmt.Sprintf("%s.conf", cfg.Name)
	f, err := os.OpenFile(path.Join(cfg.Dir, filename), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return dev.WriteConfig(f)
}
