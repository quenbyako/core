package env

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/quenbyako/core"
)

// fieldParams contains information about parsed field tags.
type fieldParams struct {
	// typ reflect.Type

	key          string
	DefaultValue string
	separator    string
	kvSeparator  string
	defaultSet   bool
	ignored      bool
}

const underscore rune = '_'

func toEnvName(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s) + 4)

	var prevClass charClass
	for i, r := range s {
		c := classOf(r)

		// Преобразуем любые "не буквы/цифры" в разделитель.
		if c == clUnderscore || c == clNone {
			if b.Len() > 0 && prevClass != clUnderscore {
				b.WriteByte('_')
				prevClass = clUnderscore
			}
			continue
		}

		// Признаки для вставки '_'
		nextLower := false
		if c.is(clUpper) && i+1 < len(s) {
			nc := rune(s[i+1])
			nextLower = unicode.IsLower(nc)
		}

		needSep := b.Len() > 0 && prevClass != clUnderscore &&
			((prevClass.is(clLower) && c.is(clUpper)) || // fooBar
				(prevClass.is(clUpper) && c.is(clUpper) && nextLower) || // HTTPServer -> HTTP_Server
				((prevClass.is(clLower | clUpper)) && c.is(clDigit)) || // Foo1
				(prevClass.is(clDigit) && (c.is(clLower | clUpper)))) // 1Foo

		if needSep {
			b.WriteByte('_')
		}

		if c.is(clDigit) {
			b.WriteRune(r)
		} else {
			b.WriteRune(unicode.ToUpper(r))
		}
		prevClass = c
	}

	return b.String()
}

type charClass uint8

const (
	clNone  charClass = 0
	clLower charClass = 1 << iota
	clUpper
	clDigit
	clUnderscore
)

func (c charClass) is(other charClass) bool { return c&other != 0 }

func classOf(r rune) charClass {
	switch {
	case r == '_':
		return clUnderscore
	case unicode.IsLower(r):
		return clLower
	case unicode.IsUpper(r):
		return clUpper
	case unicode.IsDigit(r):
		return clDigit
	default:
		return clNone
	}
}

func parseFieldParams(field reflect.StructField, prefix string) fieldParams {
	key, tags := tagOption(field.Tag.Get(tagName))
	if key == "" {
		key = toEnvName(field.Name)
	}
	if key != "-" {
		key = prefix + key
	}

	defaultValue, defaultSet := field.Tag.Lookup(tagDefault)

	separator, ok := field.Tag.Lookup(tagSeparator)
	if !ok {
		separator = ","
	}

	kvSeparator, ok := field.Tag.Lookup(tagKVSeparator)
	if !ok {
		kvSeparator = ":"
	}

	result := fieldParams{
		// typ: field.Type,

		key:          key,
		DefaultValue: defaultValue,
		separator:    separator,
		kvSeparator:  kvSeparator,

		ignored:    key == "-",
		defaultSet: defaultSet,
	}

	for _, tag := range tags {
		switch tag {
		case "":
			continue
		default:
			panic(fmt.Sprintf("%q: unsupported tag option: %q", field.Name, tag))
		}
	}

	return result
}

func setValue(ctx context.Context, v reflect.Value, p parseParams, f fieldParams, prefix string) []*FieldError {
	value, exists := p.getEnv(f.key)
	var usingDefault bool
	if !exists || value == "" {
		if !f.defaultSet {
			return []*FieldError{
				errField(p.keyWithPrefix(f.key), v.Type(), ErrValueNotSet),
			}
		}

		value = f.DefaultValue
		usingDefault = true
	}

	if v.Kind() == reflect.Pointer {
		if v.Elem().Kind() == reflect.Invalid {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	typ := v.Type() // f.typ
	parserFunc, ptrDepth, ok := core.GetParseFunc(typ)
	_ = ptrDepth // TODO: pointer restoration
	if ok {
		val, err := parserFunc(ctx, value)
		if err != nil {
			return []*FieldError{
				errField(p.keyWithPrefix(f.key), v.Type(), err),
			}
		}
		value := reflect.ValueOf(val).Convert(typ)
		v.Set(value)
		if p.onSet != nil {
			p.onSet(p.keyWithPrefix(f.key), value.Interface(), usingDefault)
		}

		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return setStruct(ctx, v, p, prefix)
	case reflect.Slice:
		return setSlice(ctx, v, value, f, p)
	case reflect.Map:
		return setMap(ctx, v, value, f, p)
	default:
		panic(fmt.Sprintf("no parser found for %v, kind %v, env var %q", v.Type().String(), v.Kind(), f.key))
	}
}

func setStruct(ctx context.Context, v reflect.Value, p parseParams, prefix string) (errs []*FieldError) {
	refType := v.Type()

	for i := 0; i < refType.NumField(); i++ {
		refField := v.Field(i)
		refTypeField := refType.Field(i)

		if err := setStructField(ctx, refField, refTypeField, p, structTagPrefix(prefix, refTypeField)); err != nil {
			errs = append(errs, err...)
		}
	}

	return errs
}

func setStructField(ctx context.Context, v reflect.Value, tags reflect.StructField, p parseParams, prefix string) []*FieldError {
	if !v.CanSet() {
		return nil
	}

	params := parseFieldParams(tags, prefix)

	if params.ignored {
		return nil
	}

	if reflect.Ptr == v.Kind() && v.Elem().Kind() == reflect.Invalid {
		v.Set(reflect.New(v.Type().Elem()))
		if v.Type().Elem().Kind() == reflect.Struct {
			v = v.Elem()
		}
	}

	if errs := setValue(ctx, v, p, params, prefix); errs != nil {
		return errs
	}

	return nil
}

func setSlice(ctx context.Context, field reflect.Value, value string, f fieldParams, p parseParams) []*FieldError {
	if field.Kind() != reflect.Slice {
		panic("field is not a slice")
	}

	itemType := field.Type().Elem()
	parserFunc, ptrDepth, ok := core.GetParseFunc(itemType)
	if !ok {
		// TODO: allow nested slices, cause in some rarest cases it may be useful
		panic(fmt.Sprintf("no parser found for %T", itemType))
	}

	parts := strings.Split(value, f.separator)

	result := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	var errs []error
	for i, part := range parts {
		r, err := parserFunc(ctx, part)
		if err != nil {
			errs = append(errs, fmt.Errorf("index %v: %w", i, err))
		}
		if len(errs) > 0 {
			// no need to continue setting values if there are errors
			continue
		}

		// pointer restoration based on ptrDepth
		v := reflect.ValueOf(r)
		for range ptrDepth {
			p := reflect.New(v.Type())
			p.Elem().Set(v)
			v = p
		}

		fmt.Println(v.Type(), field.Type().Elem(), ptrDepth)

		result.Index(i).Set(v)
	}

	var err error
	switch len(errs) {
	case 0:
		field.Set(result)
		return nil
	case 1:
		err = errs[0]
	default:
		err = errors.Join(errs...)
	}

	return []*FieldError{
		errField(p.keyWithPrefix(f.key), field.Type(), err),
	}
}

func setMap(ctx context.Context, field reflect.Value, value string, f fieldParams, p parseParams) []*FieldError {
	parts := strings.Split(value, f.separator)

	keyType := field.Type().Key()
	elemType := field.Type().Elem()

	keyParserFunc, keyPtrDepth, ok := core.GetParseFunc(keyType)
	if !ok {
		panic(fmt.Sprintf("no parser found for map key type %v", keyType))
	}
	elemParserFunc, elemPtrDepth, ok := core.GetParseFunc(elemType)
	if !ok {
		panic(fmt.Sprintf("no parser found for map elem type %v", elemType))
	}

	_, _ = keyPtrDepth, elemPtrDepth // TODO: pointer restoration

	result := reflect.MakeMapWithSize(field.Type(), len(parts))

	var errs []error
	for _, part := range parts {
		pairs := strings.SplitN(part, f.kvSeparator, 2)
		if len(pairs) != 2 {
			// direct error, no need to collect parsing errors
			return []*FieldError{
				errField(p.keyWithPrefix(f.key), field.Type(), errInvalidMapItemFormat(part, f.kvSeparator)),
			}
		}

		key, err := keyParserFunc(ctx, pairs[0])
		if err != nil {
			errs = append(errs, fmt.Errorf("key %q: %w", pairs[0], err))
			continue
		}

		elem, err := elemParserFunc(ctx, pairs[1])
		if err != nil {
			errs = append(errs, fmt.Errorf("value %q: %w", pairs[1], err))
			continue
		}

		result.SetMapIndex(reflect.ValueOf(key).Convert(keyType), reflect.ValueOf(elem).Convert(elemType))
	}

	var err error
	switch len(errs) {
	case 0:
		field.Set(result)
		return nil
	case 1:
		err = errs[0]
	default:
		err = errors.Join(errs...)
	}

	return []*FieldError{
		errField(p.keyWithPrefix(f.key), field.Type(), err),
	}
}

func tagOption(key string) (string, []string) {
	opts := strings.Split(key, ",")
	return opts[0], opts[1:]
}
