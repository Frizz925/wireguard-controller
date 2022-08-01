package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/frizz925/wireguard-controller/internal/server"
	"github.com/skip2/go-qrcode"
)

const (
	SERVER_HOST        = "dowg.kogane.moe"
	SERVER_DEVICE_NAME = "wg0"
	SERVER_LISTEN_PORT = 443

	USERS_FILE = "users.txt"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	host := SERVER_HOST
	if len(os.Args) >= 2 {
		host = os.Args[1]
	}
	srv, err := server.New(host, path.Join(cwd, "templates"))
	if err != nil {
		return err
	}
	if err := srv.Load(ctx); err != nil {
		return err
	}

	cfgDir := path.Join(cwd, "configs")
	_, err = os.Stat(cfgDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.Mkdir(cfgDir, 0700); err != nil {
			return err
		}
	}

	dev := srv.GetDevice(SERVER_DEVICE_NAME)
	if dev == nil {
		var err error
		dev, err = srv.AddDevice(ctx, SERVER_DEVICE_NAME, SERVER_LISTEN_PORT)
		if err != nil {
			return err
		}
	}

	users, err := readUsersFile(USERS_FILE)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	for _, user := range users {
		peer := dev.GetClient(user)
		if peer == nil {
			peer, err = dev.AddClient(ctx, user)
			if err != nil {
				return err
			}
		}

		buf.Reset()
		if err := peer.WriteConfig(&buf); err != nil {
			return err
		}

		prefix := path.Join(cfgDir, user)
		if err := os.WriteFile(fmt.Sprintf("%s.conf", prefix), buf.Bytes(), 0600); err != nil {
			return err
		}
		if err := qrcode.WriteFile(buf.String(), qrcode.Medium, 512, fmt.Sprintf("%s.png", prefix)); err != nil {
			return err
		}
	}

	if err := dev.WriteConfig(os.Stdout); err != nil {
		return err
	}
	return srv.Save(ctx)
}

func readUsersFile(filename string) ([]string, error) {
	f, err := os.Open(USERS_FILE)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	users := make([]string, 0)
	for sc.Scan() {
		users = append(users, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
