package secrets

import (
	"bytes"
	"context"
)

// Secret represents a retrievable opaque byte slice. Implementations may
// source data lazily (remote fetch, decryption) on each [Secret.Get] invocation.
// Callers should treat returned slices as immutable and copy when retaining.
type Secret interface {
	Get(ctx context.Context) ([]byte, error)
}

type emptySecret struct{}

var _ Secret = (*emptySecret)(nil) //nolint:grouper // type check

// NewEmptySecret returns a Secret that always yields a nil slice and no error.
// Useful for optional values where absence is not exceptional.
//
//nolint:ireturn // returns interface on intention.
func NewEmptySecret() Secret { return &emptySecret{} }

func (s *emptySecret) Get(context.Context) ([]byte, error) { return nil, nil }

type plainSecret struct {
	data []byte
}

var _ Secret = (*plainSecret)(nil) //nolint:grouper // type check

func (s *plainSecret) Get(context.Context) ([]byte, error) { return bytes.Clone(s.data), nil }

// NewPlainSecret wraps the provided bytes in a Secret performing no
// transformation.
//
//nolint:ireturn // returns interface on intention.
func NewPlainSecret(data []byte) Secret { return &plainSecret{data: data} }
