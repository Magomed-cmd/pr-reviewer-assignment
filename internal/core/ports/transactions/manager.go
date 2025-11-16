package transactions

import "context"

type Manager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type NoopManager struct{}

func (NoopManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if fn == nil {
		return nil
	}

	return fn(ctx)
}
