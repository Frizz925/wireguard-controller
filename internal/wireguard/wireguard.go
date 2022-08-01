package wireguard

import (
	"context"

	"github.com/frizz925/wireguard-controller/internal/command"
)

func Genkey(ctx context.Context) (string, error) {
	return command.OutputStringCommand(ctx, "wg", "genkey")
}

func Pubkey(ctx context.Context, privkey string) (string, error) {
	return command.InputOutputStringCommand(ctx, privkey, "wg", "pubkey")
}

func Genpsk(ctx context.Context) (string, error) {
	return command.OutputStringCommand(ctx, "wg", "genpsk")
}
