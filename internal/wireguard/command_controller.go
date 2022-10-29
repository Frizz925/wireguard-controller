package wireguard

import (
	"context"

	"github.com/frizz925/wireguard-controller/internal/commander"
)

type CommandController struct {
	*commander.Wrapper
}

func NewCommandController(cmd commander.Commander) *CommandController {
	return &CommandController{commander.NewWrapper(cmd)}
}

func (cc *CommandController) Genkey(ctx context.Context) (string, error) {
	return cc.OutputStringCommand(ctx, "wg", "genkey")
}

func (cc *CommandController) Pubkey(ctx context.Context, privkey string) (string, error) {
	return cc.InputOutputStringCommand(ctx, privkey, "wg", "pubkey")
}

func (cc *CommandController) Genpsk(ctx context.Context) (string, error) {
	return cc.OutputStringCommand(ctx, "wg", "genpsk")
}

func (cc *CommandController) Device(name string) DeviceController {
	return &CommandDeviceController{
		CommandController: cc,
		name:              name,
	}
}
