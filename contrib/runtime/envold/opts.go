package env

import (
	"fmt"
	"os"
	"reflect"
)

// newParams for the parser.
type newParams struct {
	Environment map[string]string

	// TagName specifies another tag name to use rather than the default 'env'.
	TagName string

	// PrefixTagName specifies another prefix tag name to use rather than the default 'envPrefix'.
	PrefixTagName string

	// DefaultValueTagName specifies another default tag name to use rather than
	// the default 'envDefault'.
	DefaultValueTagName string

	// RequiredIfNoDef automatically sets all fields as required if they do not
	// declare 'envDefault'.
	RequiredIfNoDef bool

	// OnSet allows to run a function when a value is set.
	OnSet OnSetFn

	// Prefix define a prefix for every key.
	Prefix string

	// UseFieldNameByDefault defines whether or not `env` should use the field
	// name by default if the `env` key is missing.
	// Note that the field name will be "converted" to conform with environment
	// variable names conventions.
	UseFieldNameByDefault bool

	// SetDefaultsForZeroValuesOnly defines whether to set defaults for zero values
	// If the `env` variable for the value is not set
	// and `envDefault` is set
	// and the value is not a zero value for the type
	// and SetDefaultsForZeroValuesOnly=true
	// the value from `envDefault` will be ignored
	// Useful for mixing default values from `envDefault` and struct initialization
	SetDefaultsForZeroValuesOnly bool

	// Custom parse functions for different types.
	FuncMap func(reflect.Type) (f ParserFunc, ptrDepth int, ok bool)

	// Used internally. maps the env variable key to its resolved string value.
	// (for env var expansion)
	rawEnvVars map[string]string
}

type NewOption func(*newParams)

func WithFuncMap(f func(reflect.Type) (f ParserFunc, ptrDepth int, ok bool)) NewOption {
	return func(p *newParams) { p.FuncMap = f }
}

func WithOnSet(onSet OnSetFn) NewOption {
	return func(p *newParams) { p.OnSet = onSet }
}

func WithTagName(name string) NewOption {
	return func(p *newParams) { p.TagName = name }
}

func WithDefaultValueTagName(name string) NewOption {
	return func(p *newParams) { p.DefaultValueTagName = name }
}

func WithRequiredIfNoDef() NewOption {
	return func(p *newParams) { p.RequiredIfNoDef = true }
}

func WithPrefixTagName(name string) NewOption {
	return func(p *newParams) { p.PrefixTagName = name }
}

func WithPrefix(prefix string) NewOption {
	return func(p *newParams) { p.Prefix = prefix }
}

func WithEnvironment(env map[string]string) NewOption {
	return func(p *newParams) { p.Environment = env }
}

func WithSetDefaultsForZeroValuesOnly() NewOption {
	return func(p *newParams) { p.SetDefaultsForZeroValuesOnly = true }
}

func WithoutSetDefaultsForZeroValuesOnly() NewOption {
	return func(p *newParams) { p.SetDefaultsForZeroValuesOnly = false }
}

func WithUseFieldNameByDefault() NewOption {
	return func(p *newParams) { p.UseFieldNameByDefault = true }
}

func defaultOptions(opts ...NewOption) newParams {
	params := newParams{
		TagName:                      "env",
		PrefixTagName:                "envPrefix",
		DefaultValueTagName:          "envDefault",
		Environment:                  ToMap(os.Environ()),
		FuncMap:                      defaultTypeParsers(),
		rawEnvVars:                   make(map[string]string),
		OnSet:                        func(string, any, bool) {},
		RequiredIfNoDef:              false,
		Prefix:                       "",
		UseFieldNameByDefault:        false,
		SetDefaultsForZeroValuesOnly: false,
	}

	for _, opt := range opts {
		opt(&params)
	}

	return params
}

func optionsWithSliceEnvPrefix(opts newParams, index int) newParams {
	return newParams{
		Environment:                  opts.Environment,
		TagName:                      opts.TagName,
		PrefixTagName:                opts.PrefixTagName,
		DefaultValueTagName:          opts.DefaultValueTagName,
		RequiredIfNoDef:              opts.RequiredIfNoDef,
		OnSet:                        opts.OnSet,
		Prefix:                       fmt.Sprintf("%s%d_", opts.Prefix, index),
		UseFieldNameByDefault:        opts.UseFieldNameByDefault,
		SetDefaultsForZeroValuesOnly: opts.SetDefaultsForZeroValuesOnly,
		FuncMap:                      opts.FuncMap,
		rawEnvVars:                   opts.rawEnvVars,
	}
}

func optionsWithEnvPrefix(field reflect.StructField, opts newParams) newParams {
	return newParams{
		Environment:                  opts.Environment,
		TagName:                      opts.TagName,
		PrefixTagName:                opts.PrefixTagName,
		DefaultValueTagName:          opts.DefaultValueTagName,
		RequiredIfNoDef:              opts.RequiredIfNoDef,
		OnSet:                        opts.OnSet,
		Prefix:                       opts.Prefix + field.Tag.Get(opts.PrefixTagName),
		UseFieldNameByDefault:        opts.UseFieldNameByDefault,
		SetDefaultsForZeroValuesOnly: opts.SetDefaultsForZeroValuesOnly,
		FuncMap:                      opts.FuncMap,
		rawEnvVars:                   opts.rawEnvVars,
	}
}
