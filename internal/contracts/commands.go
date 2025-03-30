package contracts

import "context"

type CommandsHandler interface {
	Execute(ctx context.Context, parts []string) (string, error)
}
