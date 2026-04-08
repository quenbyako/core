package env_test

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	. "github.com/quenbyako/core/contrib/runtime/envold"
)

// Basic package usage example.
func Example() {
	type Config struct {
		Foo string `env:"FOO"`
	}

	// parse:
	var cfg1 Config
	_ = Parse(&cfg1, WithEnvironment(map[string]string{"FOO": "bar"}))

	// parse with generics:
	cfg2, _ := ParseAs[Config](WithEnvironment(map[string]string{"FOO": "bar"}))

	fmt.Print(cfg1.Foo, cfg2.Foo)
	// Output: barbar
}

// Parse the environment into a struct.
func ExampleParse() {
	type Config struct {
		Home string `env:"HOME"`
	}
	var cfg Config
	if err := Parse(&cfg,
		WithEnvironment(map[string]string{
			"HOME": "/tmp/fakehome",
		}),
	); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output:  {Home:/tmp/fakehome}
}

// Parse the environment into a struct using generics.
func ExampleParseAs() {
	type Config struct {
		Home string `env:"HOME"`
	}
	cfg, err := ParseAs[Config](WithEnvironment(map[string]string{
		"HOME": "/tmp/fakehome",
	}))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output:  {Home:/tmp/fakehome}
}

func ExampleParse_required() {
	type Config struct {
		Nope string `env:"NOPE,required"`
	}
	var cfg Config
	if err := Parse(&cfg); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: env: required environment variable "NOPE" is not set
	// {Nope:}
}

// While `required` demands the environment variable to be set, it doesn't check
// its value. If you want to make sure the environment is set and not empty, you
// need to use the `notEmpty` tag option instead (`env:"SOME_ENV,notEmpty"`).
func ExampleParse_notEmpty() {
	type Config struct {
		Nope string `env:"NOPE,notEmpty"`
	}
	var cfg Config
	if err := Parse(&cfg, WithEnvironment(map[string]string{
		"NOPE": "",
	})); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: env: environment variable "NOPE" should not be empty
	// {Nope:}
}

// The `env` tag option `unset` (e.g., `env:"tagKey,unset"`) can be added
// to ensure that some environment variable is unset after reading it.
func ExampleParse_unset() {
	type Config struct {
		Secret string `env:"SECRET,unset"`
	}
	var cfg Config
	if err := Parse(&cfg, WithEnvironment(map[string]string{
		"SECRET": "1234",
	})); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v - %s", cfg, os.Getenv("SECRET"))
	// Output: {Secret:1234} -
}

// You can use `envSeparator` to define which character should be used to
// separate array items in a string.
// Similarly, you can use `envKeyValSeparator` to define which character should
// be used to separate a key from a value in a map.
// The defaults are `,` and `:`, respectively.
func ExampleParse_separator() {
	type Config struct {
		Map map[string]string `env:"CUSTOM_MAP" envSeparator:"-" envKeyValSeparator:"|"`
	}
	var cfg Config
	if err := Parse(&cfg, WithEnvironment(map[string]string{
		"CUSTOM_MAP": "k1|v1-k2|v2",
	})); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: {Map:map[k1:v1 k2:v2]}
}

// You can automatically initialize `nil` pointers regardless of if a variable
// is set for them or not.
// This behavior can be enabled by using the `init` tag option.
func ExampleParse_init() {
	type Inner struct {
		A string `env:"OLA" envDefault:"HI"`
	}
	type Config struct {
		NilInner  *Inner
		InitInner *Inner `env:",init"`
	}
	var cfg Config
	if err := Parse(&cfg); err != nil {
		fmt.Println(err)
	}
	fmt.Print(cfg.NilInner, cfg.InitInner)
	// Output: <nil> &{HI}
}

// You can define the default value for a field by either using the
// `envDefault` tag, or when initializing the `struct`.
//
// Default values defined as `struct` tags will overwrite existing values
// during `Parse`.
func ExampleParse_setDefaults() {
	type Config struct {
		Foo string `env:"DEF_FOO"`
		Bar string `env:"DEF_BAR" envDefault:"bar"`
	}
	cfg := Config{
		Foo: "foo",
	}
	if err := Parse(&cfg); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: {Foo:foo Bar:bar}
}

// You might want to listen to value sets and, for example, log something or do
// some other kind of logic.
func ExampleParse_onSet() {
	type config struct {
		Home         string `env:"HOME,required"`
		Port         int    `env:"PORT" envDefault:"3000"`
		IsProduction bool   `env:"PRODUCTION"`
		NoEnvTag     bool
		Inner        struct{} `envPrefix:"INNER_"`
	}
	var cfg config
	if err := Parse(&cfg,
		WithEnvironment(map[string]string{"HOME": "/tmp/fakehome"}),
		WithOnSet(func(tag string, value any, isDefault bool) {
			fmt.Printf("Set %s to %v (default? %v)\n", tag, value, isDefault)
		}),
	); err != nil {
		fmt.Println("failed:", err)
	}
	fmt.Printf("%+v", cfg)
	// Output: Set HOME to /tmp/fakehome (default? false)
	// Set PORT to 3000 (default? true)
	// {Home:/tmp/fakehome Port:3000 IsProduction:false NoEnvTag:false Inner:{}}
}

// By default, env supports anything that implements the `TextUnmarshaler`
// interface, which includes `time.Time`.
//
// The upside is that depending on the format you need, you don't need to change
// anything.
//
// The downside is that if you do need time in another format, you'll need to
// create your own type and implement `TextUnmarshaler`.
func ExampleParse_customTimeFormat() {
	// type MyTime time.Time
	//
	// func (t *MyTime) UnmarshalText(text []byte) error {
	// 	tt, err := time.Parse("2006-01-02", string(text))
	// 	*t = MyTime(tt)
	// 	return err
	// }

	type Config struct {
		SomeTime MyTime `env:"SOME_TIME"`
	}
	var cfg Config
	if err := Parse(&cfg, WithEnvironment(map[string]string{
		"SOME_TIME": "2021-05-06",
	})); err != nil {
		fmt.Println(err)
	}
	fmt.Print(cfg.SomeTime)
	// Output: {0 63755856000 <nil>}
}

// Parse using extra options.
func ExampleParse_customTypes() {
	// type Mapper map[reflect.Type]ParserFunc
	//
	// func (m Mapper) get(t reflect.Type) (ParserFunc, bool) {
	// 	f, ok := m[t]
	// 	return f, ok
	// }
	//
	// func UseMapper[T any](m Mapper, parseFunc func(string) (T, error)) Mapper {
	// 	typ := reflect.TypeFor[T]()
	// 	fn := func(s string) (any, error) { return parseFunc(s) }
	//
	// 	m[typ] = fn
	// 	return m
	// }

	type Thing struct {
		desc string
	}

	type Config struct {
		Thing Thing `env:"THING"`
	}

	m := Mapper{}
	m = UseMapper(m, func(v string) (Thing, error) {
		return Thing{desc: v}, nil
	})

	c := Config{}
	err := Parse(&c,
		WithEnvironment(map[string]string{"THING": "my thing"}),
		WithFuncMap(m.get),
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(c.Thing.desc)
	// Output: my thing
}

// Make all fields required by default.
func ExampleParse_allFieldsRequired() {
	type Config struct {
		Username string `env:"EX_USERNAME" envDefault:"admin"`
		Password string `env:"EX_PASSWORD"`
	}

	var cfg Config
	if err := Parse(&cfg,
		WithRequiredIfNoDef(),
	); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", cfg)
	// Output: env: required environment variable "EX_PASSWORD" is not set
	// {Username:admin Password:}
}

// Set a custom environment.
// By default, `os.Environ()` is used.
func ExampleParse_setEnv() {
	type Config struct {
		Username string `env:"EX_USERNAME" envDefault:"admin"`
		Password string `env:"EX_PASSWORD"`
	}

	var cfg Config
	if err := Parse(&cfg,
		WithEnvironment(map[string]string{
			"EX_USERNAME": "john",
			"EX_PASSWORD": "cena",
		}),
	); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", cfg)
	// Output: {Username:john Password:cena}
}

// Handling slices of complex types.
func ExampleParse_complexSlices() {
	type Test struct {
		Str string `env:"STR"`
		Num int    `env:"NUM"`
	}
	type Config struct {
		Foo []Test `envPrefix:"FOO"`
	}

	var cfg Config
	if err := Parse(&cfg, WithEnvironment(map[string]string{
		"FOO_0_STR": "a",
		"FOO_0_NUM": "1",
		"FOO_1_STR": "b",
		"FOO_1_NUM": "2",
	})); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", cfg)
	// Output: {Foo:[{Str:a Num:1} {Str:b Num:2}]}
}

// Setting prefixes for the entire config.
func ExampleParse_prefix() {
	type Config struct {
		Foo string `env:"FOO"`
	}
	var cfg Config
	if err := Parse(&cfg,
		WithEnvironment(map[string]string{"MY_APP_FOO": "a"}),
		WithPrefix("MY_APP_"),
	); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: {Foo:a}
}

// Use a different tag name than `env` and `envDefault`.
func ExampleParse_tagName() {
	type Config struct {
		Home string `json:"HOME"`
		Page string `json:"PAGE" def:"world"`
	}
	var cfg Config
	if err := Parse(&cfg,
		WithEnvironment(map[string]string{"HOME": "hello"}),
		WithTagName("json"),
		WithDefaultValueTagName("def"),
	); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: {Home:hello Page:world}
}

// If you don't want to set the `env` tag on every field, you can use the
// `UseFieldNameByDefault` option.
//
// It will use the field name to define the environment variable name.
// So, `Foo` becomes `FOO`, `FooBar` becomes `FOO_BAR`, and so on.
func ExampleParse_useFieldName() {
	type Config struct {
		Foo string
	}
	var cfg Config
	if err := Parse(&cfg,
		WithEnvironment(map[string]string{"FOO": "bar"}),
		WithUseFieldNameByDefault(),
	); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v", cfg)
	// Output: {Foo:bar}
}

func ExampleParse_errorHandling() {
	type Config struct {
		Username string `env:"EX_ERR_USERNAME" envDefault:"admin"`
		Password string `env:"EX_ERR_PASSWORD,notEmpty"`
	}

	var cfg Config
	if err := Parse(&cfg); err != nil {
		if errors.Is(err, EmptyVarError{}) {
			fmt.Println("oopsie")
		}
		aggErr := AggregateError{}
		if ok := errors.As(err, &aggErr); ok {
			for _, er := range aggErr.Errors {
				switch v := er.(type) {
				// Handle the error types you need:
				// ParseError
				// NotStructPtrError
				// NoParserError
				// NoSupportedTagOptionError
				// EnvVarIsNotSetError
				// EmptyEnvVarError
				// LoadFileContentError
				// ParseValueError
				case EmptyVarError:
					fmt.Println("daisy")
				default:
					fmt.Printf("Unknown error type %v", v)
				}
			}
		}
	}

	fmt.Printf("%+v", cfg)
	// Output: oopsie
	// daisy
	// {Username:admin Password:}
}

// You can avoid setting defaults for non zero values
// This could be useful for loading data from config file first
// and then filling the rest from env
func Example_setDefaultsForZeroValuesOnly() {
	type Config struct {
		Username string `env:"USERNAME" envDefault:"admin"`
		Password string `env:"PASSWORD" envDefault:"qwerty"`
	}

	cfg := Config{
		Username: "root",
	}

	if err := Parse(&cfg,
		WithEnvironment(map[string]string{}),
		WithSetDefaultsForZeroValuesOnly(),
	); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%+v", cfg)
	// Without SetDefaultsForZeroValuesOnly, the username would have been 'admin'.
	// Output: {Username:root Password:qwerty}
}

type Mapper map[reflect.Type]ParserFunc

func (m Mapper) get(t reflect.Type) (ParserFunc, bool) {
	f, ok := m[t]
	return f, ok
}

func UseMapper[T any](m Mapper, parseFunc func(string) (T, error)) Mapper {
	typ := reflect.TypeFor[T]()
	fn := func(s string) (any, error) { return parseFunc(s) }

	m[typ] = fn
	return m
}
