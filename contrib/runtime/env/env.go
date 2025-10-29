package env

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

const (
	tagName        = "env"
	tagPrefix      = "prefix"
	tagDefault     = "default"
	tagSeparator   = "envSeparator"
	tagKVSeparator = "envKeyValSeparator"
)

func slicePrefix(prefix string, index int) string {
	return fmt.Sprintf("%s%d_", prefix, index)
}

func structTagPrefix(prefix string, field reflect.StructField) string {
	return prefix + field.Tag.Get(tagPrefix)
}

// ParserFunc defines the signature of a function that can be used within
// `Options`' `FuncMap`.
type ParserFunc func(ctx context.Context, v string) (any, error)

// Parse parses a struct containing `env` tags and loads its values from
// environment variables.
func Parse(ctx context.Context, v any, opts ...Option) error {
	p, err := buildParseParams(opts...)
	if err != nil {
		return fmt.Errorf("options: %w", err)
	}

	return parseInternal(ctx, v, p, "")
}

// ParseAs parses the given struct type containing `env` tags and loads its
// values from environment variables.
func ParseAs[T any](ctx context.Context, opts ...Option) (T, error) {
	var t T
	if err := Parse(ctx, &t, opts...); err != nil {
		return *new(T), err
	}

	return t, nil
}

func parseInternal(ctx context.Context, v any, opts parseParams, prefix string) error {
	ptrRef := reflect.ValueOf(v)

	if ptrRef.Kind() != reflect.Ptr {
		return ErrNotStructPtr
	}

	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return ErrNotStructPtr
	}

	fields := setStruct(ctx, ref, opts, prefix)
	switch len(fields) {
	case 0:
		return nil
	case 1:
		return fields[0]
	default:
		errs := make([]error, len(fields))
		for i, f := range fields {
			errs[i] = f
		}
		return errors.Join(errs...)
	}
}
