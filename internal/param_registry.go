//nolint:forbidigo // it's necessary to use reflection here, unfortunately.
package internal

import (
	"context"
	"encoding"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"
)

//nolint:wrapcheck // errors are wrapped in parsing functions.
var (
	//nolint:exhaustive // ehmmm, no.
	defaultBuiltInParsers = map[reflect.Kind]parserFunc{ //nolint:gochecknoglobals
		reflect.Bool: func(_ context.Context, v string) (any, error) {
			return strconv.ParseBool(v)
		},
		reflect.String: func(_ context.Context, v string) (any, error) {
			return v, nil
		},
		reflect.Int: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int(i), err
		},
		reflect.Int16: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseInt(v, 10, 16)
			return int16(i), err
		},
		reflect.Int32: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int32(i), err
		},
		reflect.Int64: func(_ context.Context, v string) (any, error) {
			return strconv.ParseInt(v, 10, 64)
		},
		reflect.Int8: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseInt(v, 10, 8)
			return int8(i), err
		},
		reflect.Uint: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint(i), err
		},
		reflect.Uint16: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseUint(v, 10, 16)
			return uint16(i), err
		},
		reflect.Uint32: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint32(i), err
		},
		reflect.Uint64: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseUint(v, 10, 64)
			return i, err
		},
		reflect.Uint8: func(_ context.Context, v string) (any, error) {
			i, err := strconv.ParseUint(v, 10, 8)
			return uint8(i), err
		},
		reflect.Float64: func(_ context.Context, v string) (any, error) {
			return strconv.ParseFloat(v, 64)
		},
		reflect.Float32: func(_ context.Context, v string) (any, error) {
			f, err := strconv.ParseFloat(v, 32)
			return float32(f), err
		},
	}
	envRegistry = map[reflect.Type]parserFunc{
		reflect.TypeFor[slog.Level]():    parseLogLevel,
		reflect.TypeFor[url.URL]():       parseURL,
		reflect.TypeFor[time.Duration](): parseDuration,
		reflect.TypeFor[time.Location](): parseLocation,
		reflect.TypeFor[*os.File]():      nil, // TODO: implement that
		reflect.TypeFor[fs.File]():       nil, // TODO: implement that
	}
)

func RegisterEnvParser[T any](parseFunc func(context.Context, string) (T, error)) {
	typ := reflect.TypeFor[T]()
	if _, exists := envRegistry[typ]; exists {
		panic(fmt.Sprintf("parser for %v already registered", typ))
	}

	envRegistry[typ] = func(ctx context.Context, v string) (any, error) { return parseFunc(ctx, v) }
}

type parserFunc = func(context.Context, string) (any, error)

// Deprecated: This is a temporary function to aid migration. Use [GetParseFunc] instead.
func GetAllParseFunc() map[reflect.Type]parserFunc { return envRegistry }

func GetParseFunc(typ reflect.Type) (f parserFunc, ptrDepth int, ok bool) {
	// unpacking pointers
	inner := typ
	depth := 0

	for {
		if f, ok := envRegistry[typ]; ok {
			return f, depth, true
		}

		if f, ok := handleTextUnmarshaler(typ); ok {
			return f, depth, true
		}

		if f, ok := defaultBuiltInParsers[typ.Kind()]; ok {
			return f, depth, true
		}

		if inner.Kind() != reflect.Pointer {
			break
		}

		typ = typ.Elem()
		depth++
	}

	return nil, 0, false
}

func handleTextUnmarshaler(typ reflect.Type) (parserFunc, bool) {
	unmarshalTyp := reflect.TypeFor[encoding.TextUnmarshaler]()

	// unpack pointers
	base := typ
	depth := 0

	for base.Kind() == reflect.Pointer {
		base = base.Elem()
		depth++
	}

	switch {
	case base.Implements(unmarshalTyp): // func (t T) UnmarshalText(...)
		return valueTextUnmarshaler(base, depth), true

	case reflect.PointerTo(base).Implements(unmarshalTyp): // func (t *T) UnmarshalText(...)
		return ptrTextUnmarshaler(base, depth), true

	default:
		return nil, false
	}
}

func valueTextUnmarshaler(base reflect.Type, depth int) parserFunc {
	return func(_ context.Context, raw string) (any, error) {
		value := reflect.New(base).Elem() // value T

		unmarshaler, ok := value.Interface().(encoding.TextUnmarshaler)
		if !ok {
			// assertion checks above by calling [reflect.Type.Implements].
			panic("unreachable") //nolint:forbidigo // unreachable
		}

		if err := unmarshaler.UnmarshalText([]byte(raw)); err != nil {
			return nil, ErrUnmarshalFunc(value.Type(), err)
		}

		if depth == 0 {
			return value.Interface(), nil
		}
		// Reversing depth back.
		out := value
		for range depth {
			p := reflect.New(out.Type())
			p.Elem().Set(out)
			out = p
		}

		return out.Interface(), nil
	}
}

func ptrTextUnmarshaler(base reflect.Type, depth int) parserFunc {
	return func(_ context.Context, raw string) (any, error) {
		value := reflect.New(base) // *T

		unmarshaler, ok := value.Interface().(encoding.TextUnmarshaler)
		if !ok {
			// assertion checks above by calling [reflect.Type.Implements].
			panic("unreachable") //nolint:forbidigo // unreachable
		}

		if err := unmarshaler.UnmarshalText([]byte(raw)); err != nil {
			return nil, ErrUnmarshalFunc(value.Type(), err)
		}

		if depth == 0 {
			return value.Elem().Interface(), nil // typ == T
		}

		if depth == 1 {
			return value.Interface(), nil // typ == *T
		}

		// need to add **T, ***T, ...
		out := value
		for i := 1; i < depth; i++ {
			w := reflect.New(out.Type()) // new(*T), new(**T), ...
			w.Elem().Set(out)
			out = w
		}

		return out.Interface(), nil
	}
}
