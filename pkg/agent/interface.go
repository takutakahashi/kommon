package agent

import "context"

type Agent interface {
	Execute(ctx context.Context, input string) (string, error)
}
