package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/frizz925/wireguard-controller/internal/commander"
	"github.com/frizz925/wireguard-controller/internal/config"
	"github.com/frizz925/wireguard-controller/internal/device"
	"github.com/frizz925/wireguard-controller/internal/logger"
	"github.com/frizz925/wireguard-controller/internal/server"
	"github.com/frizz925/wireguard-controller/internal/wireguard"
	"github.com/melbahja/goph"
	"github.com/skip2/go-qrcode"
	"gopkg.in/yaml.v3"

	clientRepoPkg "github.com/frizz925/wireguard-controller/internal/repositories/client"
	serverRepoPkg "github.com/frizz925/wireguard-controller/internal/repositories/server"
)

var deviceRegex = regexp.MustCompile("^[a-z0-9]+$")

type serverConfig struct {
	config.Server

	Host string
	Cwd  string
	Dir  string

	ServerRepo serverRepoPkg.Repository
	ClientRepo clientRepoPkg.Repository
	Logger     *logger.Logger
}

type sshConfig struct {
	config.SSH
	Logger *logger.Logger
}

type deviceConfig struct {
	config.Device
	Server *server.Server

	Host string
	Name string
	Dir  string

	Controller wireguard.DeviceController
	Logger     *logger.Logger
}

type clientConfig struct {
	config.User
	Device *device.ServerDevice

	FilePrefix string

	Buffer *bytes.Buffer
	Logger *logger.Logger
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	var err error
	serverRepo := serverRepoPkg.NewLocalRepository()
	clientRepo := clientRepoPkg.NewLocalRepository()
	log := logger.New(os.Stderr)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfgDir := path.Join(cwd, "configs")

	hosts := os.Args[1:]
	if len(hosts) <= 0 {
		hosts, err = readHostDirs(cfgDir)
		if err != nil {
			return err
		}
	}

	for _, host := range hosts {
		hostDir := path.Join(cfgDir, host)
		fi, err := os.Stat(hostDir)
		if err != nil {
			return err
		} else if !fi.IsDir() {
			return fmt.Errorf("%s is not a directory", host)
		}

		f, err := os.Open(path.Join(hostDir, "server.yaml"))
		if err != nil {
			return err
		}
		defer f.Close()
		log.Log("Host %s", host)

		var srv config.Server
		if err := yaml.NewDecoder(f).Decode(&srv); err != nil {
			return err
		}
		scfg := &serverConfig{
			Server:     srv,
			Host:       host,
			Cwd:        cwd,
			Dir:        hostDir,
			ServerRepo: serverRepo,
			ClientRepo: clientRepo,
			Logger:     log.Indent(),
		}
		if err := generateServer(ctx, scfg); err != nil {
			return err
		}
	}
	return nil
}

func generateServer(ctx context.Context, cfg *serverConfig) error {
	log := cfg.Logger

	sshHost := cfg.SSH.Hostname
	if sshHost == "" {
		sshHost = cfg.Host
	}
	log.Log("Connection %s (SSH)", sshHost)
	client, err := connectSSH(sshHost, &sshConfig{
		SSH:    cfg.SSH,
		Logger: log.Indent(),
	})
	if err != nil {
		return err
	}

	cmd := commander.NewSSHCommander(client)
	ctrl := wireguard.NewCommandController(cmd)

	srv, err := server.New(&server.Config{
		Host:         cfg.Host,
		TemplatesDir: path.Join(cfg.Cwd, "templates"),
		Controller:   ctrl,
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
		file := path.Base(filePath)
		if file == "server.yaml" {
			continue
		}

		idx := strings.Index(file, ".yaml")
		if idx <= 0 {
			continue
		}
		name := file[:idx]
		if err := validateDeviceName(name); err != nil {
			return err
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		log.Log("Device %s", name)

		var dev config.Device
		if err := yaml.NewDecoder(f).Decode(&dev); err != nil {
			return err
		}

		dcfg := &deviceConfig{
			Device:     dev,
			Server:     srv,
			Host:       cfg.Host,
			Name:       name,
			Dir:        path.Join(cfg.Dir, name),
			Controller: ctrl.Device(name),
			Logger:     log.Indent(),
		}
		if err := generateDevice(ctx, dcfg); err != nil {
			return err
		}
	}
	return nil
}

func generateDevice(ctx context.Context, cfg *deviceConfig) error {
	var err error
	srv, ctrl, log := cfg.Server, cfg.Controller, cfg.Logger

	dev := srv.GetDevice(cfg.Name)
	if dev == nil {
		dev, err = srv.AddDevice(ctx, cfg.Name, cfg.Device)
		if err != nil {
			return err
		}
		log.Log("Device created")
	} else {
		dev.Apply(cfg.Device)
		log.Log("Device updated")
	}

	// Create directories
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
	userMap := make(map[string]bool)
	for _, user := range cfg.Users {
		log.Log("Client %s (%s)", user.Name, user.Address)
		ccfg := &clientConfig{
			User:       user,
			Device:     dev,
			FilePrefix: path.Join(cfg.Dir, user.Name),
			Buffer:     &buf,
			Logger:     log.Indent(),
		}
		if err := generateClient(ctx, ccfg); err != nil {
			return err
		}
		userMap[user.Name] = true
	}

	// Check for removed user
	for _, user := range dev.GetClientNames() {
		if _, ok := userMap[user]; ok {
			continue
		}
		peer, err := dev.RemoveClient(ctx, user)
		if err != nil {
			return err
		}
		log.Log("Client %s deleted", peer.Name)
	}

	if err := srv.Save(ctx); err != nil {
		return err
	}

	buf.Reset()
	if err := dev.WriteConfig(&buf); err != nil {
		return err
	}
	if err := ctrl.SaveConfig(ctx, buf.Bytes()); err != nil {
		return err
	}
	log.Log("Device config created")

	enabled, err := ctrl.IsEnabled(ctx)
	if err != nil {
		return err
	} else if !enabled {
		if err := ctrl.Enable(ctx); err != nil {
			return err
		}
		log.Log("Device enabled")
		return nil
	}

	active, err := ctrl.IsActive(ctx)
	if err != nil {
		return err
	} else if active {
		if err := ctrl.Restart(ctx); err != nil {
			return err
		}
		log.Log("Device restarted")
	} else {
		if err := ctrl.Start(ctx); err != nil {
			return err
		}
		log.Log("Device started")
	}
	return nil
}

func generateClient(ctx context.Context, cfg *clientConfig) error {
	var err error
	dev, log := cfg.Device, cfg.Logger

	peer := dev.GetClient(cfg.Name)
	if peer == nil {
		peer, err = dev.AddClient(ctx, cfg.User)
		if err != nil {
			return err
		}
		log.Log("Client created")
	} else {
		peer.Apply(cfg.User)
		log.Log("Client updated")
	}

	buf := cfg.Buffer
	buf.Reset()
	if err := peer.WriteConfig(buf); err != nil {
		return err
	}

	prefix := cfg.FilePrefix
	if err := os.WriteFile(fmt.Sprintf("%s.conf", prefix), buf.Bytes(), 0600); err != nil {
		return err
	}
	log.Log("Client config created")
	if err := qrcode.WriteFile(buf.String(), qrcode.Medium, 512, fmt.Sprintf("%s.png", prefix)); err != nil {
		return err
	}
	log.Log("Client QR config created")
	return nil
}

func readHostDirs(cfgDir string) ([]string, error) {
	dirs, err := os.ReadDir(cfgDir)
	if err != nil {
		return nil, err
	}

	hosts := make([]string, 0)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		hosts = append(hosts, dir.Name())
	}
	return hosts, nil
}

func connectSSH(host string, cfg *sshConfig) (*goph.Client, error) {
	log := cfg.Logger
	auth, err := goph.Key(cfg.IdentityFile, cfg.Passphrase)
	if err != nil {
		return nil, err
	}
	log.Log("Connection establishing")
	client, err := goph.New(cfg.User, host, auth)
	if err != nil {
		return nil, err
	}
	log.Log("Connection established")
	return client, nil
}

func validateDeviceName(name string) error {
	if deviceRegex.MatchString(name) {
		return nil
	}
	return fmt.Errorf("device name should be lowercase alphanumeric: %s", name)
}
