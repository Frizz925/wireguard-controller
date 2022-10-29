package wireguard

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

type CommandDeviceController struct {
	*CommandController
	name string
}

func (cdc *CommandDeviceController) Name() string {
	return cdc.name
}

func (cdc *CommandDeviceController) ServiceName() string {
	return fmt.Sprintf("wg-quick@%s", cdc.name)
}

func (cdc *CommandDeviceController) SaveConfig(ctx context.Context, content []byte) error {
	confPath := fmt.Sprintf("/etc/wireguard/%s.conf", cdc.Name())
	if err := cdc.sudo(ctx, "install", "-m", "600", "/dev/null", confPath); err != nil {
		return err
	}
	return cdc.sudoInput(ctx, bytes.NewReader(content), "sudo", "tee", confPath)
}

func (cdc *CommandDeviceController) IsEnabled(ctx context.Context) (bool, error) {
	res, err := cdc.sudoOutputString(ctx, "systemctl", "is-enabled", cdc.ServiceName())
	if err != nil {
		return false, err
	}
	return res == "enabled", nil
}

func (cdc *CommandDeviceController) IsActive(ctx context.Context) (bool, error) {
	res, err := cdc.sudoOutputString(ctx, "systemctl", "is-active", cdc.ServiceName())
	if err != nil {
		return false, err
	}
	return res == "active", nil
}

func (cdc *CommandDeviceController) Enable(ctx context.Context) error {
	return cdc.sudo(ctx, "systemctl", "enable", "--now", cdc.ServiceName())
}

func (cdc *CommandDeviceController) Start(ctx context.Context) error {
	return cdc.sudo(ctx, "systemctl", "start", cdc.ServiceName())
}

func (cdc *CommandDeviceController) Restart(ctx context.Context) error {
	return cdc.sudo(ctx, "systemctl", "restart", cdc.ServiceName())
}

func (cdc *CommandDeviceController) sudo(ctx context.Context, name string, args ...string) error {
	args = append([]string{name}, args...)
	return cdc.SimpleCommand(ctx, "sudo", args...)
}

func (cdc *CommandDeviceController) sudoInput(ctx context.Context, input io.Reader, name string, args ...string) error {
	args = append([]string{name}, args...)
	return cdc.InputCommand(ctx, input, "sudo", args...)
}

func (cdc *CommandDeviceController) sudoOutputString(ctx context.Context, name string, args ...string) (string, error) {
	args = append([]string{name}, args...)
	return cdc.OutputStringCommand(ctx, "sudo", args...)
}
