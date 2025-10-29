package secrets

import (
	"context"
	"fmt"
	"io"
)

// Engine provides access to secrets addressed by a storage-specific string.
// Implementations may perform remote lookups, caching, or decryption. Errors
// should be descriptive; callers can wrap them but generally do not assume
// sentinel types besides those defined in this package.
type Engine interface {
	io.Closer
	GetSecret(ctx context.Context, addr string) (Secret, error)
}

type unsetStorage struct {
	name string
}

var _ Engine = (*unsetStorage)(nil) //nolint:grouper // type check

// NewUnsetStorage returns an [Engine] that always fails lookups indicating the
// named storage is not configured. Useful as a placeholder for unconfigured
// backends.
//
//nolint:ireturn // returns interface on intention.
func NewUnsetStorage(name string) Engine { return &unsetStorage{name: name} }

// GetSecret always returns a [Secret] wrapping the constant data.
//
//nolint:ireturn // returns interface on intention.
func (u *unsetStorage) GetSecret(ctx context.Context, key string) (Secret, error) {
	return nil, fmt.Errorf("storage %q is unset", u.name)
}

func (u *unsetStorage) Close() error {
	return nil
}

type constantStorage struct {
	data []byte
}

var _ Engine = (*constantStorage)(nil) //nolint:grouper // type check

// NewConstantStorage constructs an [Engine] serving a single immutable secret
// value. This is handy for tests or embedding small static configuration.
//
//nolint:ireturn // returns interface on intention.
func NewConstantStorage(data []byte) Engine {
	return &constantStorage{data: data}
}

// GetSecret always returns a [Secret] wrapping the constant data.
//
//nolint:ireturn // returns interface on intention.
func (c *constantStorage) GetSecret(context.Context, string) (Secret, error) {
	return NewPlainSecret(c.data), nil
}

func (c *constantStorage) Close() error { return nil }
