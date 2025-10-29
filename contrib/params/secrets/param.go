package secrets

import (
	"context"
	"fmt"

	"github.com/quenbyako/core"
	"github.com/quenbyako/core/secrets"
)

type Secret interface {
	Get(ctx context.Context) ([]byte, error)
}

func init() { //nolint:gochecknoinits // there is no other way to register parsers
	core.RegisterEnvParser(parseSecret)
}

type rawSecret struct {
	wrapped secrets.Secret
	path    string
}

var (
	_ Secret        = (*rawSecret)(nil)
	_ core.EnvParam = (*rawSecret)(nil)
)

//nolint:ireturn // returns interface on intention.
func parseSecret(ctx context.Context, v string) (Secret, error) {
	return &rawSecret{
		wrapped: nil, // will be initialized later
		path:    v,
	}, nil
}

func (s *rawSecret) Get(ctx context.Context) ([]byte, error) {
	if s.wrapped == nil {
		panic("uninitialized") //nolint:forbidigo // unreachable
	}

	res, err := s.wrapped.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting secret: %w", err)
	}

	return res, nil
}

func (s *rawSecret) Configure(ctx context.Context, data *core.ConfigureData) error {
	if data.Secrets == nil {
		return secrets.ErrEngineNotConfigured
	}

	secret, err := data.Secrets.GetSecret(ctx, s.path)
	if err != nil {
		return fmt.Errorf("getting secret %q: %w", s.path, err)
	}

	s.wrapped = secret

	return nil
}

func (s *rawSecret) Acquire(ctx context.Context, data *core.AcquireData) error   { return nil }
func (s *rawSecret) Shutdown(ctx context.Context, data *core.ShutdownData) error { return nil }
