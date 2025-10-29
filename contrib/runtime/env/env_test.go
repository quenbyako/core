package env_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	// "path/filepath"
	"reflect"
	// "runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/quenbyako/core/contrib/runtime/env"
)

func tos(v ...any) string {
	ss := []string{}
	for _, s := range v {
		ss = append(ss, fmt.Sprintf("%v", s))
	}
	return strings.Join(ss, ",")
}

type unmarshaler struct {
	time.Duration
}

// TextUnmarshaler implements encoding.TextUnmarshaler.
func (d *unmarshaler) UnmarshalText(data []byte) (err error) {
	if len(data) != 0 {
		d.Duration, err = time.ParseDuration(string(data))
	} else {
		d.Duration = 0
	}
	return err
}

type Config struct { //nolint:maligned
	String     string    `env:"STRING"`
	StringPtr  *string   `env:"STRING"`
	Strings    []string  `env:"STRINGS"`
	StringPtrs []*string `env:"STRINGS"`

	Bool     bool    `env:"BOOL"`
	BoolPtr  *bool   `env:"BOOL"`
	Bools    []bool  `env:"BOOLS"`
	BoolPtrs []*bool `env:"BOOLS"`

	Int     int    `env:"INT"`
	IntPtr  *int   `env:"INT"`
	Ints    []int  `env:"INTS"`
	IntPtrs []*int `env:"INTS"`

	Int8     int8    `env:"INT8"`
	Int8Ptr  *int8   `env:"INT8"`
	Int8s    []int8  `env:"INT8S"`
	Int8Ptrs []*int8 `env:"INT8S"`

	Int16     int16    `env:"INT16"`
	Int16Ptr  *int16   `env:"INT16"`
	Int16s    []int16  `env:"INT16S"`
	Int16Ptrs []*int16 `env:"INT16S"`

	Int32     int32    `env:"INT32"`
	Int32Ptr  *int32   `env:"INT32"`
	Int32s    []int32  `env:"INT32S"`
	Int32Ptrs []*int32 `env:"INT32S"`

	Int64     int64    `env:"INT64"`
	Int64Ptr  *int64   `env:"INT64"`
	Int64s    []int64  `env:"INT64S"`
	Int64Ptrs []*int64 `env:"INT64S"`

	Uint     uint    `env:"UINT"`
	UintPtr  *uint   `env:"UINT"`
	Uints    []uint  `env:"UINTS"`
	UintPtrs []*uint `env:"UINTS"`

	Uint8     uint8    `env:"UINT8"`
	Uint8Ptr  *uint8   `env:"UINT8"`
	Uint8s    []uint8  `env:"UINT8S"`
	Uint8Ptrs []*uint8 `env:"UINT8S"`

	Uint16     uint16    `env:"UINT16"`
	Uint16Ptr  *uint16   `env:"UINT16"`
	Uint16s    []uint16  `env:"UINT16S"`
	Uint16Ptrs []*uint16 `env:"UINT16S"`

	Uint32     uint32    `env:"UINT32"`
	Uint32Ptr  *uint32   `env:"UINT32"`
	Uint32s    []uint32  `env:"UINT32S"`
	Uint32Ptrs []*uint32 `env:"UINT32S"`

	Uint64     uint64    `env:"UINT64"`
	Uint64Ptr  *uint64   `env:"UINT64"`
	Uint64s    []uint64  `env:"UINT64S"`
	Uint64Ptrs []*uint64 `env:"UINT64S"`

	Float32     float32    `env:"FLOAT32"`
	Float32Ptr  *float32   `env:"FLOAT32"`
	Float32s    []float32  `env:"FLOAT32S"`
	Float32Ptrs []*float32 `env:"FLOAT32S"`

	Float64     float64    `env:"FLOAT64"`
	Float64Ptr  *float64   `env:"FLOAT64"`
	Float64s    []float64  `env:"FLOAT64S"`
	Float64Ptrs []*float64 `env:"FLOAT64S"`

	Duration     time.Duration    `env:"DURATION"`
	Durations    []time.Duration  `env:"DURATIONS"`
	DurationPtr  *time.Duration   `env:"DURATION"`
	DurationPtrs []*time.Duration `env:"DURATIONS"`

	Location     time.Location    `env:"LOCATION"`
	Locations    []time.Location  `env:"LOCATIONS"`
	LocationPtr  *time.Location   `env:"LOCATION"`
	LocationPtrs []*time.Location `env:"LOCATIONS"`

	Unmarshaler     unmarshaler    `env:"UNMARSHALER"`
	UnmarshalerPtr  *unmarshaler   `env:"UNMARSHALER"`
	Unmarshalers    []unmarshaler  `env:"UNMARSHALERS"`
	UnmarshalerPtrs []*unmarshaler `env:"UNMARSHALERS"`

	URL     url.URL    `env:"URL"`
	URLPtr  *url.URL   `env:"URL"`
	URLs    []url.URL  `env:"URLS"`
	URLPtrs []*url.URL `env:"URLS"`

	StringWithDefault string `env:"DATABASE_URL" envDefault:"postgres://localhost:5432/db"`

	CustomSeparator []string `env:"SEPSTRINGS" envSeparator:":"`

	NonDefined struct {
		String string `env:"NONDEFINED_STR"`
	}

	NestedNonDefined struct {
		NonDefined struct {
			String string `env:"STR"`
		} `envPrefix:"NONDEFINED_"`
	} `envPrefix:"PRF_"`

	NotAnEnv   string
	unexported string `env:"FOO"`
}

type ParentStruct struct {
	InnerStruct    *InnerStruct `env:",init"`
	NilInnerStruct *InnerStruct
	unexported     *InnerStruct
	Ignored        *http.Client
}

type InnerStruct struct {
	Inner  string `env:"innervar"`
	Number uint   `env:"innernum"`
}

type ForNestedStruct struct {
	NestedStruct
}

type NestedStruct struct {
	NestedVar string `env:"nestedvar"`
}

func TestIssue245(t *testing.T) {
	t.Setenv("NAME_NOT_SET", "")
	type user struct {
		Name string `env:"NAME_NOT_SET" envDefault:"abcd"`
	}
	cfg := user{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, cfg.Name, "abcd")
}

func TestParsesEnv(t *testing.T) {
	str1 := "str1"
	str2 := "str2"

	bool1 := true
	bool2 := false

	int1 := -1
	int2 := 2

	var int81 int8 = -2
	var int82 int8 = 5
	var int161 int16 = -24
	var int162 int16 = 15
	var int321 int32 = -14
	var int322 int32 = 154
	var int641 int64 = -12
	var int642 int64 = 150
	var uint1 uint = 1
	var uint2 uint = 2
	var uint81 uint8 = 15
	var uint82 uint8 = 51
	var uint161 uint16 = 532
	var uint162 uint16 = 123
	var uint321 uint32 = 93
	var uint322 uint32 = 14
	var uint641 uint64 = 5
	var uint642 uint64 = 43
	var float321 float32 = 9.3
	var float322 float32 = 1.1
	float641 := 1.53
	float642 := 0.5
	duration1 := time.Second
	duration2 := time.Second * 4
	location1 := time.UTC
	location2, errLoadLocation := time.LoadLocation("Europe/Berlin")
	isNoErr(t, errLoadLocation)
	unmarshaler1 := unmarshaler{time.Minute}
	unmarshaler2 := unmarshaler{time.Millisecond * 1232}
	url1 := "https://goreleaser.com"
	url2 := "https://caarlos0.dev"
	nonDefinedStr := "nonDefinedStr"

	m := map[string]string{
		"STRING":             tos(str1),
		"STRINGS":            tos(str1, str2),
		"BOOL":               tos(bool1),
		"BOOLS":              tos(bool1, bool2),
		"INT":                tos(int1),
		"INTS":               tos(int1, int2),
		"INT8":               tos(int81),
		"INT8S":              tos(int81, int82),
		"INT16":              tos(int161),
		"INT16S":             tos(int161, int162),
		"INT32":              tos(int321),
		"INT32S":             tos(int321, int322),
		"INT64":              tos(int641),
		"INT64S":             tos(int641, int642),
		"UINT":               tos(uint1),
		"UINTS":              tos(uint1, uint2),
		"UINT8":              tos(uint81),
		"UINT8S":             tos(uint81, uint82),
		"UINT16":             tos(uint161),
		"UINT16S":            tos(uint161, uint162),
		"UINT32":             tos(uint321),
		"UINT32S":            tos(uint321, uint322),
		"UINT64":             tos(uint641),
		"UINT64S":            tos(uint641, uint642),
		"FLOAT32":            tos(float321),
		"FLOAT32S":           tos(float321, float322),
		"FLOAT64":            tos(float641),
		"FLOAT64S":           tos(float641, float642),
		"DURATION":           tos(duration1),
		"DURATIONS":          tos(duration1, duration2),
		"LOCATION":           tos(location1),
		"LOCATIONS":          tos(location1, location2),
		"UNMARSHALER":        tos(unmarshaler1.Duration),
		"UNMARSHALERS":       tos(unmarshaler1.Duration, unmarshaler2.Duration),
		"URL":                tos(url1),
		"URLS":               tos(url1, url2),
		"SEPSTRINGS":         str1 + ":" + str2,
		"NONDEFINED_STR":     nonDefinedStr,
		"PRF_NONDEFINED_STR": nonDefinedStr,
		"FOO":                str1,
	}

	cfg := Config{}
	isNoErr(t, Parse(t.Context(), &cfg, WithEnvironment(m)))

	isEqual(t, str1, cfg.String)
	isEqual(t, &str1, cfg.StringPtr)
	isEqual(t, str1, cfg.Strings[0])
	isEqual(t, str2, cfg.Strings[1])
	isEqual(t, &str1, cfg.StringPtrs[0])
	isEqual(t, &str2, cfg.StringPtrs[1])

	isEqual(t, bool1, cfg.Bool)
	isEqual(t, &bool1, cfg.BoolPtr)
	isEqual(t, bool1, cfg.Bools[0])
	isEqual(t, bool2, cfg.Bools[1])
	isEqual(t, &bool1, cfg.BoolPtrs[0])
	isEqual(t, &bool2, cfg.BoolPtrs[1])

	isEqual(t, int1, cfg.Int)
	isEqual(t, &int1, cfg.IntPtr)
	isEqual(t, int1, cfg.Ints[0])
	isEqual(t, int2, cfg.Ints[1])
	isEqual(t, &int1, cfg.IntPtrs[0])
	isEqual(t, &int2, cfg.IntPtrs[1])

	isEqual(t, int81, cfg.Int8)
	isEqual(t, &int81, cfg.Int8Ptr)
	isEqual(t, int81, cfg.Int8s[0])
	isEqual(t, int82, cfg.Int8s[1])
	isEqual(t, &int81, cfg.Int8Ptrs[0])
	isEqual(t, &int82, cfg.Int8Ptrs[1])

	isEqual(t, int161, cfg.Int16)
	isEqual(t, &int161, cfg.Int16Ptr)
	isEqual(t, int161, cfg.Int16s[0])
	isEqual(t, int162, cfg.Int16s[1])
	isEqual(t, &int161, cfg.Int16Ptrs[0])
	isEqual(t, &int162, cfg.Int16Ptrs[1])

	isEqual(t, int321, cfg.Int32)
	isEqual(t, &int321, cfg.Int32Ptr)
	isEqual(t, int321, cfg.Int32s[0])
	isEqual(t, int322, cfg.Int32s[1])
	isEqual(t, &int321, cfg.Int32Ptrs[0])
	isEqual(t, &int322, cfg.Int32Ptrs[1])

	isEqual(t, int641, cfg.Int64)
	isEqual(t, &int641, cfg.Int64Ptr)
	isEqual(t, int641, cfg.Int64s[0])
	isEqual(t, int642, cfg.Int64s[1])
	isEqual(t, &int641, cfg.Int64Ptrs[0])
	isEqual(t, &int642, cfg.Int64Ptrs[1])

	isEqual(t, uint1, cfg.Uint)
	isEqual(t, &uint1, cfg.UintPtr)
	isEqual(t, uint1, cfg.Uints[0])
	isEqual(t, uint2, cfg.Uints[1])
	isEqual(t, &uint1, cfg.UintPtrs[0])
	isEqual(t, &uint2, cfg.UintPtrs[1])

	isEqual(t, uint81, cfg.Uint8)
	isEqual(t, &uint81, cfg.Uint8Ptr)
	isEqual(t, uint81, cfg.Uint8s[0])
	isEqual(t, uint82, cfg.Uint8s[1])
	isEqual(t, &uint81, cfg.Uint8Ptrs[0])
	isEqual(t, &uint82, cfg.Uint8Ptrs[1])

	isEqual(t, uint161, cfg.Uint16)
	isEqual(t, &uint161, cfg.Uint16Ptr)
	isEqual(t, uint161, cfg.Uint16s[0])
	isEqual(t, uint162, cfg.Uint16s[1])
	isEqual(t, &uint161, cfg.Uint16Ptrs[0])
	isEqual(t, &uint162, cfg.Uint16Ptrs[1])

	isEqual(t, uint321, cfg.Uint32)
	isEqual(t, &uint321, cfg.Uint32Ptr)
	isEqual(t, uint321, cfg.Uint32s[0])
	isEqual(t, uint322, cfg.Uint32s[1])
	isEqual(t, &uint321, cfg.Uint32Ptrs[0])
	isEqual(t, &uint322, cfg.Uint32Ptrs[1])

	isEqual(t, uint641, cfg.Uint64)
	isEqual(t, &uint641, cfg.Uint64Ptr)
	isEqual(t, uint641, cfg.Uint64s[0])
	isEqual(t, uint642, cfg.Uint64s[1])
	isEqual(t, &uint641, cfg.Uint64Ptrs[0])
	isEqual(t, &uint642, cfg.Uint64Ptrs[1])

	isEqual(t, float321, cfg.Float32)
	isEqual(t, &float321, cfg.Float32Ptr)
	isEqual(t, float321, cfg.Float32s[0])
	isEqual(t, float322, cfg.Float32s[1])
	isEqual(t, &float321, cfg.Float32Ptrs[0])

	isEqual(t, float641, cfg.Float64)
	isEqual(t, &float641, cfg.Float64Ptr)
	isEqual(t, float641, cfg.Float64s[0])
	isEqual(t, float642, cfg.Float64s[1])
	isEqual(t, &float641, cfg.Float64Ptrs[0])
	isEqual(t, &float642, cfg.Float64Ptrs[1])

	isEqual(t, duration1, cfg.Duration)
	isEqual(t, &duration1, cfg.DurationPtr)
	isEqual(t, duration1, cfg.Durations[0])
	isEqual(t, duration2, cfg.Durations[1])
	isEqual(t, &duration1, cfg.DurationPtrs[0])
	isEqual(t, &duration2, cfg.DurationPtrs[1])

	isEqual(t, *location1, cfg.Location)
	isEqual(t, location1, cfg.LocationPtr)
	isEqual(t, *location1, cfg.Locations[0])
	isEqual(t, *location2, cfg.Locations[1])
	isEqual(t, location1, cfg.LocationPtrs[0])
	isEqual(t, location2, cfg.LocationPtrs[1])

	isEqual(t, unmarshaler1, cfg.Unmarshaler)
	isEqual(t, &unmarshaler1, cfg.UnmarshalerPtr)
	isEqual(t, unmarshaler1, cfg.Unmarshalers[0])
	isEqual(t, unmarshaler2, cfg.Unmarshalers[1])
	isEqual(t, &unmarshaler1, cfg.UnmarshalerPtrs[0])
	isEqual(t, &unmarshaler2, cfg.UnmarshalerPtrs[1])

	isEqual(t, url1, cfg.URL.String())
	isEqual(t, url1, cfg.URLPtr.String())
	isEqual(t, url1, cfg.URLs[0].String())
	isEqual(t, url2, cfg.URLs[1].String())
	isEqual(t, url1, cfg.URLPtrs[0].String())
	isEqual(t, url2, cfg.URLPtrs[1].String())

	isEqual(t, "postgres://localhost:5432/db", cfg.StringWithDefault)
	isEqual(t, nonDefinedStr, cfg.NonDefined.String)
	isEqual(t, nonDefinedStr, cfg.NestedNonDefined.NonDefined.String)

	isEqual(t, str1, cfg.CustomSeparator[0])
	isEqual(t, str2, cfg.CustomSeparator[1])

	isEqual(t, cfg.NotAnEnv, "")

	isEqual(t, cfg.unexported, "")
}

func TestParsesEnv_Map(t *testing.T) {
	type config struct {
		MapStringString                map[string]string `env:"MAP_STRING_STRING" envSeparator:","`
		MapStringInt64                 map[string]int64  `env:"MAP_STRING_INT64"`
		MapStringBool                  map[string]bool   `env:"MAP_STRING_BOOL" envSeparator:";"`
		CustomSeparatorMapStringString map[string]string `env:"CUSTOM_SEPARATOR_MAP_STRING_STRING" envSeparator:"," envKeyValSeparator:"|"`
	}

	mss := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	t.Setenv("MAP_STRING_STRING", "k1:v1,k2:v2")

	msi := map[string]int64{
		"k1": 1,
		"k2": 2,
	}
	t.Setenv("MAP_STRING_INT64", "k1:1,k2:2")

	msb := map[string]bool{
		"k1": true,
		"k2": false,
	}
	t.Setenv("MAP_STRING_BOOL", "k1:true;k2:false")

	withCustomSeparator := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	t.Setenv("CUSTOM_SEPARATOR_MAP_STRING_STRING", "k1|v1,k2|v2")

	var cfg config
	isNoErr(t, Parse(t.Context(), &cfg))

	isEqual(t, mss, cfg.MapStringString)
	isEqual(t, msi, cfg.MapStringInt64)
	isEqual(t, msb, cfg.MapStringBool)
	isEqual(t, withCustomSeparator, cfg.CustomSeparatorMapStringString)
}

func TestParsesEnvInvalidMap(t *testing.T) {
	type config struct {
		MapStringString map[string]string `env:"MAP_STRING_STRING" envSeparator:","`
	}

	var cfg config
	err := Parse(t.Context(), &cfg, WithEnvironment(map[string]string{
		"MAP_STRING_STRING": "k1,k2:v2",
	}))
	e := &InvalidMapItemFormatError{}
	isTrue(t, errors.As(err, &e))
	isEqual(t, "k1", e.Item)
	isEqual(t, ":", e.KVSeparator)
}

/*
func TestParseCustomMapType(t *testing.T) {
	type custommap map[string]bool

	type config struct {
		SecretKey custommap `env:"SECRET_KEY"`
	}

	t.Setenv("SECRET_KEY", "somesecretkey:1")

	var cfg config
	isNoErr(t, Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(custommap{}): func(_ string) (any, error) {
			return custommap(map[string]bool{}), nil
		},
	}}))
}

func TestParseMapCustomKeyType(t *testing.T) {
	type CustomKey string

	type config struct {
		SecretKey map[CustomKey]bool `env:"SECRET"`
	}

	t.Setenv("SECRET", "somesecretkey:1")

	var cfg config
	isNoErr(t, Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(CustomKey("")): func(value string) (any, error) {
			return CustomKey(value), nil
		},
	}}))
}

func TestParseMapCustomKeyNoParser(t *testing.T) {
	type CustomKey struct{}

	type config struct {
		SecretKey map[CustomKey]bool `env:"SECRET"`
	}

	t.Setenv("SECRET", "somesecretkey:1")

	var cfg config
	err := Parse(t.Context(), &cfg)
	isTrue(t, errors.Is(err, NoParserError{}))
}

func TestParseMapCustomValueNoParser(t *testing.T) {
	type Customval struct{}

	type config struct {
		SecretKey map[string]Customval `env:"SECRET"`
	}

	t.Setenv("SECRET", "somesecretkey:1")

	var cfg config
	err := Parse(t.Context(), &cfg)
	isTrue(t, errors.Is(err, NoParserError{}))
}

func TestParseMapCustomKeyTypeError(t *testing.T) {
	type CustomKey string

	type config struct {
		SecretKey map[CustomKey]bool `env:"SECRET"`
	}

	t.Setenv("SECRET", "somesecretkey:1")

	var cfg config
	err := Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(CustomKey("")): func(_ string) (any, error) {
			return nil, fmt.Errorf("custom error")
		},
	}})
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParseMapCustomValueTypeError(t *testing.T) {
	type Customval string

	type config struct {
		SecretKey map[string]Customval `env:"SECRET"`
	}

	t.Setenv("SECRET", "somesecretkey:1")

	var cfg config
	err := Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(Customval("")): func(_ string) (any, error) {
			return nil, fmt.Errorf("custom error")
		},
	}})
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestSetenvAndTagOptsChain(t *testing.T) {
	type config struct {
		Key1 string `mytag:"KEY1,required"`
		Key2 int    `mytag:"KEY2,required"`
	}
	envs := map[string]string{
		"KEY1": "VALUE1",
		"KEY2": "3",
	}

	cfg := config{}
	isNoErr(t, Parse(t.Context(), &cfg, Options{TagName: "mytag", Environment: envs}))
	isEqual(t, "VALUE1", cfg.Key1)
	isEqual(t, 3, cfg.Key2)
}

func TestJSONTag(t *testing.T) {
	type config struct {
		Key1 string `json:"KEY1"`
		Key2 int    `json:"KEY2"`
	}

	t.Setenv("KEY1", "VALUE7")
	t.Setenv("KEY2", "5")

	cfg := config{}
	isNoErr(t, Parse(t.Context(), &cfg, Options{TagName: "json"}))
	isEqual(t, "VALUE7", cfg.Key1)
	isEqual(t, 5, cfg.Key2)
}

func TestParsesEnvInner(t *testing.T) {
	t.Setenv("innervar", "someinnervalue")
	t.Setenv("innernum", "8")
	cfg := ParentStruct{
		InnerStruct: &InnerStruct{},
		unexported:  &InnerStruct{},
	}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "someinnervalue", cfg.InnerStruct.Inner)
	isEqual(t, uint(8), cfg.InnerStruct.Number)
}

func TestParsesEnvInner_WhenInnerStructPointerIsNil(t *testing.T) {
	t.Setenv("innervar", "someinnervalue")
	t.Setenv("innernum", "8")
	cfg := ParentStruct{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "someinnervalue", cfg.InnerStruct.Inner)
	isEqual(t, uint(8), cfg.InnerStruct.Number)
}

func TestParsesEnvInnerFails(t *testing.T) {
	type config struct {
		Foo struct {
			Number int `env:"NUMBER"`
		}
	}
	t.Setenv("NUMBER", "not-a-number")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "Number" of type "int": strconv.ParseInt: parsing "not-a-number": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParsesEnvInnerFailsMultipleErrors(t *testing.T) {
	type config struct {
		Foo struct {
			Name   string `env:"NAME,required"`
			Number int    `env:"NUMBER"`
			Bar    struct {
				Age int `env:"AGE,required"`
			}
		}
	}
	t.Setenv("NUMBER", "not-a-number")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: required environment variable "NAME" is not set; parse error on field "Number" of type "int": strconv.ParseInt: parsing "not-a-number": invalid syntax; required environment variable "AGE" is not set`)
	isTrue(t, errors.Is(err, ParseError{}))
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
}

func TestParsesEnvInnerNil(t *testing.T) {
	t.Setenv("innervar", "someinnervalue")
	cfg := ParentStruct{}
	isNoErr(t, Parse(t.Context(), &cfg))
}

func TestParsesEnvInnerInvalid(t *testing.T) {
	t.Setenv("innernum", "-547")
	cfg := ParentStruct{
		InnerStruct: &InnerStruct{},
	}
	err := Parse(t.Context(), &cfg)
	isErrorWithMessage(t, err, `env: parse error on field "Number" of type "uint": strconv.ParseUint: parsing "-547": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParsesEnvNested(t *testing.T) {
	t.Setenv("nestedvar", "somenestedvalue")
	var cfg ForNestedStruct
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "somenestedvalue", cfg.NestedVar)
}

func TestEmptyVars(t *testing.T) {
	cfg := Config{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "", cfg.String)
	isEqual(t, false, cfg.Bool)
	isEqual(t, 0, cfg.Int)
	isEqual(t, uint(0), cfg.Uint)
	isEqual(t, uint64(0), cfg.Uint64)
	isEqual(t, int64(0), cfg.Int64)
	isEqual(t, 0, len(cfg.Strings))
	isEqual(t, 0, len(cfg.CustomSeparator))
	isEqual(t, 0, len(cfg.Ints))
	isEqual(t, 0, len(cfg.Bools))
}

func TestPassAnInvalidPtr(t *testing.T) {
	var thisShouldBreak int
	err := Parse(t.Context(), &thisShouldBreak)
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

func TestPassReference(t *testing.T) {
	cfg := Config{}
	err := Parse(t.Context(), cfg)
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

func TestInvalidBool(t *testing.T) {
	t.Setenv("BOOL", "should-be-a-bool")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Bool" of type "bool": strconv.ParseBool: parsing "should-be-a-bool": invalid syntax; parse error on field "BoolPtr" of type "*bool": strconv.ParseBool: parsing "should-be-a-bool": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidInt(t *testing.T) {
	t.Setenv("INT", "should-be-an-int")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Int" of type "int": strconv.ParseInt: parsing "should-be-an-int": invalid syntax; parse error on field "IntPtr" of type "*int": strconv.ParseInt: parsing "should-be-an-int": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidUint(t *testing.T) {
	t.Setenv("UINT", "-44")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Uint" of type "uint": strconv.ParseUint: parsing "-44": invalid syntax; parse error on field "UintPtr" of type "*uint": strconv.ParseUint: parsing "-44": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat32(t *testing.T) {
	t.Setenv("FLOAT32", "AAA")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Float32" of type "float32": strconv.ParseFloat: parsing "AAA": invalid syntax; parse error on field "Float32Ptr" of type "*float32": strconv.ParseFloat: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat64(t *testing.T) {
	t.Setenv("FLOAT64", "AAA")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Float64" of type "float64": strconv.ParseFloat: parsing "AAA": invalid syntax; parse error on field "Float64Ptr" of type "*float64": strconv.ParseFloat: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidUint64(t *testing.T) {
	t.Setenv("UINT64", "AAA")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Uint64" of type "uint64": strconv.ParseUint: parsing "AAA": invalid syntax; parse error on field "Uint64Ptr" of type "*uint64": strconv.ParseUint: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidInt64(t *testing.T) {
	t.Setenv("INT64", "AAA")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Int64" of type "int64": strconv.ParseInt: parsing "AAA": invalid syntax; parse error on field "Int64Ptr" of type "*int64": strconv.ParseInt: parsing "AAA": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidInt64Slice(t *testing.T) {
	t.Setenv("BADINTS", "A,2,3")
	type config struct {
		BadFloats []int64 `env:"BADINTS"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]int64": strconv.ParseInt: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidUInt64Slice(t *testing.T) {
	t.Setenv("BADINTS", "A,2,3")
	type config struct {
		BadFloats []uint64 `env:"BADINTS"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]uint64": strconv.ParseUint: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat32Slice(t *testing.T) {
	t.Setenv("BADFLOATS", "A,2.0,3.0")
	type config struct {
		BadFloats []float32 `env:"BADFLOATS"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]float32": strconv.ParseFloat: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidFloat64Slice(t *testing.T) {
	t.Setenv("BADFLOATS", "A,2.0,3.0")
	type config struct {
		BadFloats []float64 `env:"BADFLOATS"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadFloats" of type "[]float64": strconv.ParseFloat: parsing "A": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidBoolsSlice(t *testing.T) {
	t.Setenv("BADBOOLS", "t,f,TRUE,faaaalse")
	type config struct {
		BadBools []bool `env:"BADBOOLS"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "BadBools" of type "[]bool": strconv.ParseBool: parsing "faaaalse": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidDuration(t *testing.T) {
	t.Setenv("DURATION", "should-be-a-valid-duration")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Duration" of type "time.Duration": unable to parse duration: time: invalid duration "should-be-a-valid-duration"; parse error on field "DurationPtr" of type "*time.Duration": unable to parse duration: time: invalid duration "should-be-a-valid-duration"`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidDurations(t *testing.T) {
	t.Setenv("DURATIONS", "1s,contains-an-invalid-duration,3s")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Durations" of type "[]time.Duration": unable to parse duration: time: invalid duration "contains-an-invalid-duration"; parse error on field "DurationPtrs" of type "[]*time.Duration": unable to parse duration: time: invalid duration "contains-an-invalid-duration"`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidLocation(t *testing.T) {
	t.Setenv("LOCATION", "should-be-a-valid-location")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Location" of type "time.Location": unable to parse location: unknown time zone should-be-a-valid-location; parse error on field "LocationPtr" of type "*time.Location": unable to parse location: unknown time zone should-be-a-valid-location`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestInvalidLocations(t *testing.T) {
	t.Setenv("LOCATIONS", "should-be-a-valid-location,UTC,Europe/Berlin")
	err := Parse(t.Context(), &Config{})
	isErrorWithMessage(t, err, `env: parse error on field "Locations" of type "[]time.Location": unable to parse location: unknown time zone should-be-a-valid-location; parse error on field "LocationPtrs" of type "[]*time.Location": unable to parse location: unknown time zone should-be-a-valid-location`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParseStructWithoutEnvTag(t *testing.T) {
	cfg := Config{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, cfg.NotAnEnv, "")
}

func TestParseStructWithInvalidFieldKind(t *testing.T) {
	type config struct {
		WontWorkByte byte `env:"BLAH"`
	}
	t.Setenv("BLAH", "a")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "WontWorkByte" of type "uint8": strconv.ParseUint: parsing "a": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestUnsupportedSliceType(t *testing.T) {
	type config struct {
		WontWork []map[int]int `env:"WONTWORK"`
	}

	t.Setenv("WONTWORK", "1,2,3")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: no parser found for field "WontWork" of type "[]map[int]int"`)
	isTrue(t, errors.Is(err, NoParserError{}))
}

func TestBadSeparator(t *testing.T) {
	type config struct {
		WontWork []int `env:"WONTWORK" envSeparator:":"`
	}

	t.Setenv("WONTWORK", "1,2,3,4")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "WontWork" of type "[]int": strconv.ParseInt: parsing "1,2,3,4": invalid syntax`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestNoErrorRequiredSet(t *testing.T) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required"`
	}

	cfg := &config{}

	t.Setenv("IS_REQUIRED", "")
	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "", cfg.IsRequired)
}

func TestHook(t *testing.T) {
	type config struct {
		Something string `env:"SOMETHING" envDefault:"important"`
		Another   string `env:"ANOTHER"`
		Nope      string
		Inner     struct{} `envPrefix:"FOO_"`
	}

	cfg := &config{}
	t.Setenv("ANOTHER", "1")

	type onSetArgs struct {
		tag       string
		key       any
		isDefault bool
	}

	var onSetCalled []onSetArgs

	isNoErr(t, Parse(t.Context(), cfg, Options{
		OnSet: func(tag string, value any, isDefault bool) {
			onSetCalled = append(onSetCalled, onSetArgs{tag, value, isDefault})
		},
	}))
	isEqual(t, "important", cfg.Something)
	isEqual(t, "1", cfg.Another)
	isEqual(t, 2, len(onSetCalled))
	isEqual(t, onSetArgs{"SOMETHING", "important", true}, onSetCalled[0])
	isEqual(t, onSetArgs{"ANOTHER", "1", false}, onSetCalled[1])
}

func TestErrorRequiredWithDefault(t *testing.T) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required" envDefault:"important"`
	}

	cfg := &config{}

	t.Setenv("IS_REQUIRED", "")
	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "important", cfg.IsRequired)
}

func TestErrorRequiredNotSet(t *testing.T) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: required environment variable "IS_REQUIRED" is not set`)
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
}

func TestNoErrorNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "1")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,notEmpty"`
	}
	isNoErr(t, Parse(t.Context(), &config{}))
}

func TestNoErrorRequiredAndNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "1")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required,notEmpty"`
	}
	isNoErr(t, Parse(t.Context(), &config{}))
}

func TestErrorNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,notEmpty"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: environment variable "IS_REQUIRED" should not be empty`)
	isTrue(t, errors.Is(err, EmptyVarError{}))
}

func TestErrorRequiredAndNotEmptySet(t *testing.T) {
	t.Setenv("IS_REQUIRED", "")
	type config struct {
		IsRequired string `env:"IS_REQUIRED,notEmpty,required"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: environment variable "IS_REQUIRED" should not be empty`)
	isTrue(t, errors.Is(err, EmptyVarError{}))
}

func TestErrorRequiredNotSetWithDefault(t *testing.T) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required" envDefault:"important"`
	}

	cfg := &config{}
	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "important", cfg.IsRequired)
}

func TestParseExpandOption(t *testing.T) {
	type config struct {
		Host        string `env:"HOST" envDefault:"localhost"`
		Port        int    `env:"PORT,expand" envDefault:"3000"`
		SecretKey   string `env:"SECRET_KEY,expand"`
		ExpandKey   string `env:"EXPAND_KEY"`
		CompoundKey string `env:"HOST_PORT,expand" envDefault:"${HOST}:${PORT}"`
		Default     string `env:"DEFAULT,expand" envDefault:"def1"`
	}

	t.Setenv("HOST", "localhost")
	t.Setenv("PORT", "3000")
	t.Setenv("EXPAND_KEY", "qwerty12345")
	t.Setenv("SECRET_KEY", "${EXPAND_KEY}")

	cfg := config{}
	err := Parse(t.Context(), &cfg)

	isNoErr(t, err)
	isEqual(t, "localhost", cfg.Host)
	isEqual(t, 3000, cfg.Port)
	isEqual(t, "qwerty12345", cfg.SecretKey)
	isEqual(t, "qwerty12345", cfg.ExpandKey)
	isEqual(t, "localhost:3000", cfg.CompoundKey)
	isEqual(t, "def1", cfg.Default)
}

func TestParseExpandWithDefaultOption(t *testing.T) {
	type config struct {
		Host            string `env:"HOST" envDefault:"localhost"`
		Port            int    `env:"PORT,expand" envDefault:"3000"`
		OtherPort       int    `env:"OTHER_PORT" envDefault:"4000"`
		CompoundDefault string `env:"HOST_PORT,expand" envDefault:"${HOST}:${PORT}"`
		SimpleDefault   string `env:"DEFAULT,expand" envDefault:"def1"`
		MixedDefault    string `env:"MIXED_DEFAULT,expand" envDefault:"$USER@${HOST}:${OTHER_PORT}"`
		OverrideDefault string `env:"OVERRIDE_DEFAULT,expand"`
		DefaultIsExpand string `env:"DEFAULT_IS_EXPAND,expand" envDefault:"$THIS_IS_EXPAND"`
		NoDefault       string `env:"NO_DEFAULT,expand"`
	}

	t.Setenv("OTHER_PORT", "5000")
	t.Setenv("USER", "jhon")
	t.Setenv("THIS_IS_USED", "this is used instead")
	t.Setenv("OVERRIDE_DEFAULT", "msg: ${THIS_IS_USED}")
	t.Setenv("THIS_IS_EXPAND", "msg: ${THIS_IS_USED}")
	t.Setenv("NO_DEFAULT", "$PORT:$OTHER_PORT")

	cfg := config{}
	err := Parse(t.Context(), &cfg)

	isNoErr(t, err)
	isEqual(t, "localhost", cfg.Host)
	isEqual(t, 3000, cfg.Port)
	isEqual(t, 5000, cfg.OtherPort)
	isEqual(t, "localhost:3000", cfg.CompoundDefault)
	isEqual(t, "def1", cfg.SimpleDefault)
	isEqual(t, "jhon@localhost:5000", cfg.MixedDefault)
	isEqual(t, "msg: this is used instead", cfg.OverrideDefault)
	isEqual(t, "3000:5000", cfg.NoDefault)
}

func TestParseUnsetRequireOptions(t *testing.T) {
	type config struct {
		Password string `env:"PASSWORD,unset,required"`
	}
	cfg := config{}

	err := Parse(t.Context(), &cfg)
	isErrorWithMessage(t, err, `env: required environment variable "PASSWORD" is not set`)
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
	t.Setenv("PASSWORD", "superSecret")
	isNoErr(t, Parse(t.Context(), &cfg))

	isEqual(t, "superSecret", cfg.Password)
	unset, exists := os.LookupEnv("PASSWORD")
	isEqual(t, "", unset)
	isEqual(t, false, exists)
}

func TestCustomParser(t *testing.T) {
	type foo struct {
		name string
	}

	type bar struct {
		Name string `env:"OTHER_CUSTOM"`
		Foo  *foo   `env:"BLAH_CUSTOM"`
	}

	type config struct {
		Var   foo  `env:"VAR_CUSTOM"`
		Foo   *foo `env:"BLAH_CUSTOM"`
		Other *bar
	}

	t.Setenv("VAR_CUSTOM", "test")
	t.Setenv("OTHER_CUSTOM", "test2")
	t.Setenv("BLAH_CUSTOM", "test3")

	runtest := func(t *testing.T) {
		t.Helper()
		cfg := &config{
			Other: &bar{},
		}
		err := Parse(t.Context(), cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
			reflect.TypeOf(foo{}): func(v string) (any, error) {
				return foo{name: v}, nil
			},
		}})

		isNoErr(t, err)
		isEqual(t, cfg.Var.name, "test")
		isEqual(t, cfg.Foo.name, "test3")
		isEqual(t, cfg.Other.Name, "test2")
		isEqual(t, cfg.Other.Foo.name, "test3")
	}

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("%d", i), runtest)
	}
}

func TestIssue226(t *testing.T) {
	type config struct {
		Inner struct {
			Abc []byte `env:"ABC" envDefault:"asdasd"`
			Def []byte `env:"DEF" envDefault:"a"`
		}
		Hij []byte `env:"HIJ"`
		Lmn []byte `env:"LMN"`
	}

	t.Setenv("HIJ", "a")
	t.Setenv("LMN", "b")

	cfg := &config{}
	isNoErr(t, Parse(t.Context(), cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf([]byte{0}): func(v string) (any, error) {
			if v == "a" {
				return []byte("nope"), nil
			}
			return []byte(v), nil
		},
	}}))
	isEqual(t, cfg.Inner.Abc, []byte("asdasd"))
	isEqual(t, cfg.Inner.Def, []byte("nope"))
	isEqual(t, cfg.Hij, []byte("nope"))
	isEqual(t, cfg.Lmn, []byte("b"))
}

func TestParseNoPtr(t *testing.T) {
	type foo struct{}
	err := Parse(t.Context(), foo{}, Options{})
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

func TestParseInvalidType(t *testing.T) {
	var c int
	err := Parse(t.Context(), &c, Options{})
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

func TestCustomParserError(t *testing.T) {
	type foo struct {
		name string
	}

	customParserFunc := func(_ string) (any, error) {
		return nil, errors.New("something broke")
	}

	t.Run("single", func(t *testing.T) {
		type config struct {
			Var foo `env:"VAR"`
		}

		t.Setenv("VAR", "single")
		cfg := &config{}
		err := Parse(t.Context(), cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
			reflect.TypeOf(foo{}): customParserFunc,
		}})

		isEqual(t, cfg.Var.name, "")
		isErrorWithMessage(t, err, `env: parse error on field "Var" of type "env.foo": something broke`)
		isTrue(t, errors.Is(err, ParseError{}))
	})

	t.Run("slice", func(t *testing.T) {
		type config struct {
			Var []foo `env:"VAR2"`
		}
		t.Setenv("VAR2", "slice,slace")

		cfg := &config{}
		err := Parse(t.Context(), cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
			reflect.TypeOf(foo{}): customParserFunc,
		}})

		isEqual(t, cfg.Var, nil)
		isErrorWithMessage(t, err, `env: parse error on field "Var" of type "[]env.foo": something broke`)
		isTrue(t, errors.Is(err, ParseError{}))
	})
}

func TestCustomParserBasicType(t *testing.T) {
	type ConstT int32

	type config struct {
		Const ConstT `env:"CONST_"`
	}

	exp := ConstT(123)
	t.Setenv("CONST_", fmt.Sprintf("%d", exp))

	customParserFunc := func(v string) (any, error) {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		r := ConstT(i)
		return r, nil
	}

	cfg := &config{}
	err := Parse(t.Context(), cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(ConstT(0)): customParserFunc,
	}})

	isNoErr(t, err)
	isEqual(t, exp, cfg.Const)
}

func TestCustomParserUint64Alias(t *testing.T) {
	type T uint64

	var one T = 1

	type config struct {
		Val T `env:"" envDefault:"1x"`
	}

	parserCalled := false

	tParser := func(value string) (any, error) {
		parserCalled = true
		trimmed := strings.TrimSuffix(value, "x")
		i, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, err
		}
		return T(i), nil
	}

	cfg := config{}

	err := Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(one): tParser,
	}})

	isTrue(t, parserCalled)
	isNoErr(t, err)
	isEqual(t, T(1), cfg.Val)
}

func TestTypeCustomParserBasicInvalid(t *testing.T) {
	type ConstT int32

	type config struct {
		Const ConstT `env:"CONST_"`
	}

	t.Setenv("CONST_", "foobar")

	customParserFunc := func(_ string) (any, error) {
		return nil, errors.New("random error")
	}

	cfg := &config{}
	err := Parse(t.Context(), cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(ConstT(0)): customParserFunc,
	}})

	isEqual(t, cfg.Const, ConstT(0))
	isErrorWithMessage(t, err, `env: parse error on field "Const" of type "env.ConstT": random error`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestCustomParserNotCalledForNonAlias(t *testing.T) {
	type T uint64
	type U uint64

	type config struct {
		Val   uint64 `env:"" envDefault:"33"`
		Other U      `env:"OTHER_NAME" envDefault:"44"`
	}

	tParserCalled := false

	tParser := func(_ string) (any, error) {
		tParserCalled = true
		return T(99), nil
	}

	cfg := config{}

	err := Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(T(0)): tParser,
	}})

	isFalse(t, tParserCalled)
	isNoErr(t, err)
	isEqual(t, uint64(33), cfg.Val)
	isEqual(t, U(44), cfg.Other)
}

func TestCustomParserBasicUnsupported(t *testing.T) {
	type ConstT struct {
		A int
	}

	type config struct {
		Const ConstT `env:"CONST_"`
	}

	t.Setenv("CONST_", "42")

	cfg := &config{}
	err := Parse(t.Context(), cfg)

	isEqual(t, cfg.Const, ConstT{0})
	isErrorWithMessage(t, err, `env: no parser found for field "Const" of type "env.ConstT"`)
	isTrue(t, errors.Is(err, NoParserError{}))
}

func TestUnsupportedStructType(t *testing.T) {
	type config struct {
		Foo http.Client `env:"FOO"`
	}
	t.Setenv("FOO", "foo")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: no parser found for field "Foo" of type "http.Client"`)
	isTrue(t, errors.Is(err, NoParserError{}))
}

func TestEmptyOption(t *testing.T) {
	type config struct {
		Var string `env:"VAR,"`
	}

	cfg := &config{}

	t.Setenv("VAR", "")
	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "", cfg.Var)
}

func TestErrorOptionNotRecognized(t *testing.T) {
	type config struct {
		Var string `env:"VAR,not_supported!"`
	}
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: tag option "not_supported!" not supported`)
	isTrue(t, errors.Is(err, NoSupportedTagOptionError{}))
}

func TestTextUnmarshalerError(t *testing.T) {
	type config struct {
		Unmarshaler unmarshaler `env:"UNMARSHALER"`
	}
	t.Setenv("UNMARSHALER", "invalid")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "Unmarshaler" of type "env.unmarshaler": time: invalid duration "invalid"`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestTextUnmarshalersError(t *testing.T) {
	type config struct {
		Unmarshalers []unmarshaler `env:"UNMARSHALERS"`
	}
	t.Setenv("UNMARSHALERS", "1s,invalid")
	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "Unmarshalers" of type "[]env.unmarshaler": time: invalid duration "invalid"`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestParseURL(t *testing.T) {
	type config struct {
		ExampleURL url.URL `env:"EXAMPLE_URL" envDefault:"https://google.com"`
	}
	var cfg config
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "https://google.com", cfg.ExampleURL.String())
}

func TestParseInvalidURL(t *testing.T) {
	type config struct {
		ExampleURL url.URL `env:"EXAMPLE_URL_2"`
	}
	t.Setenv("EXAMPLE_URL_2", "nope://s s/")

	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: parse error on field "ExampleURL" of type "url.URL": unable to parse URL: parse "nope://s s/": invalid character " " in host name`)
	isTrue(t, errors.Is(err, ParseError{}))
}

func TestIgnoresUnexported(t *testing.T) {
	type unexportedConfig struct {
		home  string `env:"HOME"`
		Home2 string `env:"HOME"`
	}
	cfg := unexportedConfig{}

	t.Setenv("HOME", "/tmp/fakehome")
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, cfg.home, "")
	isEqual(t, "/tmp/fakehome", cfg.Home2)
}

type LogLevel int8

func (l *LogLevel) UnmarshalText(text []byte) error {
	txt := string(text)
	switch txt {
	case "debug":
		*l = DebugLevel
	case "info":
		*l = InfoLevel
	default:
		return fmt.Errorf("unknown level: %q", txt)
	}

	return nil
}

const (
	DebugLevel LogLevel = iota - 1
	InfoLevel
)

func TestPrecedenceUnmarshalText(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_LEVELS", "debug,info")

	type config struct {
		LogLevel  LogLevel   `env:"LOG_LEVEL"`
		LogLevels []LogLevel `env:"LOG_LEVELS"`
	}
	var cfg config

	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, DebugLevel, cfg.LogLevel)
	isEqual(t, []LogLevel{DebugLevel, InfoLevel}, cfg.LogLevels)
}

func TestFile(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file"`
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "sec_key")
	isNoErr(t, os.WriteFile(file, []byte("secret"), 0o660))

	t.Setenv("SECRET_KEY", file)

	cfg := config{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "secret", cfg.SecretKey)
}

func TestFileNoParam(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file"`
	}

	cfg := config{}
	isNoErr(t, Parse(t.Context(), &cfg))
}

func TestFileNoParamRequired(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file,required"`
	}

	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, `env: required environment variable "SECRET_KEY" is not set`)
	isTrue(t, errors.Is(err, VarIsNotSetError{}))
}

func TestFileBadFile(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file"`
	}

	filename := "not-a-real-file"
	t.Setenv("SECRET_KEY", filename)

	oserr := "no such file or directory"
	if runtime.GOOS == "windows" {
		oserr = "The system cannot find the file specified."
	}

	err := Parse(t.Context(), &config{})
	isErrorWithMessage(t, err, fmt.Sprintf("env: could not load content of file %q from variable SECRET_KEY: open %s: %s", filename, filename, oserr))
	isTrue(t, errors.Is(err, LoadFileContentError{}))
}

func TestFileWithDefault(t *testing.T) {
	type config struct {
		SecretKey string `env:"SECRET_KEY,file,expand" envDefault:"${FILE}"`
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "sec_key")
	isNoErr(t, os.WriteFile(file, []byte("secret"), 0o660))

	t.Setenv("FILE", file)

	cfg := config{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "secret", cfg.SecretKey)
}

func TestCustomSliceType(t *testing.T) {
	type customslice []byte

	type config struct {
		SecretKey customslice `env:"SECRET_KEY"`
	}

	t.Setenv("SECRET_KEY", "somesecretkey")

	var cfg config
	isNoErr(t, Parse(t.Context(), &cfg, Options{FuncMap: map[reflect.Type]ParserFunc{
		reflect.TypeOf(customslice{}): func(value string) (any, error) {
			return customslice(value), nil
		},
	}}))
}

type MyTime time.Time

func (t *MyTime) UnmarshalText(text []byte) error {
	tt, err := time.Parse(t.Context(), "2006-01-02", string(text))
	*t = MyTime(tt)
	return err
}

func TestCustomTimeParser(t *testing.T) {
	type config struct {
		SomeTime MyTime `env:"SOME_TIME"`
	}

	t.Setenv("SOME_TIME", "2021-05-06")

	var cfg config
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, 2021, time.Time(cfg.SomeTime).Year())
	isEqual(t, time.Month(5), time.Time(cfg.SomeTime).Month())
	isEqual(t, 6, time.Time(cfg.SomeTime).Day())
}

func TestRequiredIfNoDefOption(t *testing.T) {
	type Tree struct {
		Fruit string `env:"FRUIT"`
	}
	type config struct {
		Name  string `env:"NAME"`
		Genre string `env:"GENRE" envDefault:"Unknown"`
		Tree
	}
	var cfg config

	t.Run("missing", func(t *testing.T) {
		err := Parse(t.Context(), &cfg, Options{RequiredIfNoDef: true})
		isErrorWithMessage(t, err, `env: required environment variable "NAME" is not set; required environment variable "FRUIT" is not set`)
		isTrue(t, errors.Is(err, VarIsNotSetError{}))
		t.Setenv("NAME", "John")
		err = Parse(t.Context(), &cfg, Options{RequiredIfNoDef: true})
		isErrorWithMessage(t, err, `env: required environment variable "FRUIT" is not set`)
		isTrue(t, errors.Is(err, VarIsNotSetError{}))
	})

	t.Run("all set", func(t *testing.T) {
		t.Setenv("NAME", "John")
		t.Setenv("FRUIT", "Apple")

		// should not trigger an error for the missing 'GENRE' env because it has a default value.
		isNoErr(t, Parse(t.Context(), &cfg, Options{RequiredIfNoDef: true}))
	})
}

func TestRequiredIfNoDefNested(t *testing.T) {
	type Server struct {
		Host string `env:"HOST"`
		Port uint16 `env:"PORT"`
	}
	type API struct {
		Server
		Token string `env:"TOKEN"`
	}
	type config struct {
		API API `envPrefix:"SERVER_"`
	}

	t.Run("missing", func(t *testing.T) {
		var cfg config
		t.Setenv("SERVER_HOST", "https://google.com")
		t.Setenv("SERVER_TOKEN", "0xdeadfood")

		err := Parse(t.Context(), &cfg, Options{RequiredIfNoDef: true})
		isErrorWithMessage(t, err, `env: required environment variable "SERVER_PORT" is not set`)
		isTrue(t, errors.Is(err, VarIsNotSetError{}))
	})

	t.Run("all set", func(t *testing.T) {
		var cfg config
		t.Setenv("SERVER_HOST", "https://google.com")
		t.Setenv("SERVER_PORT", "443")
		t.Setenv("SERVER_TOKEN", "0xdeadfood")

		isNoErr(t, Parse(t.Context(), &cfg, Options{RequiredIfNoDef: true}))
	})
}

func TestPrefix(t *testing.T) {
	type Config struct {
		Home string `env:"HOME"`
	}
	type ComplexConfig struct {
		Foo   Config `envPrefix:"FOO_"`
		Bar   Config `envPrefix:"BAR_"`
		Clean Config
	}
	cfg := ComplexConfig{}
	isNoErr(t, Parse(t.Context(), &cfg, Options{Environment: map[string]string{"FOO_HOME": "/foo", "BAR_HOME": "/bar", "HOME": "/clean"}}))
	isEqual(t, "/foo", cfg.Foo.Home)
	isEqual(t, "/bar", cfg.Bar.Home)
	isEqual(t, "/clean", cfg.Clean.Home)
}

func TestPrefixPointers(t *testing.T) {
	type Test struct {
		Str string `env:"TEST"`
	}
	type ComplexConfig struct {
		Foo   *Test `envPrefix:"FOO_"`
		Bar   *Test `envPrefix:"BAR_"`
		Clean *Test
	}

	cfg := ComplexConfig{
		Foo:   &Test{},
		Bar:   &Test{},
		Clean: &Test{},
	}
	isNoErr(t, Parse(t.Context(), &cfg, Options{Environment: map[string]string{"FOO_TEST": "kek", "BAR_TEST": "lel", "TEST": "clean"}}))
	isEqual(t, "kek", cfg.Foo.Str)
	isEqual(t, "lel", cfg.Bar.Str)
	isEqual(t, "clean", cfg.Clean.Str)
}

func TestNestedPrefixPointer(t *testing.T) {
	type ComplexConfig struct {
		Foo struct {
			Str string `env:"STR"`
		} `envPrefix:"FOO_"`
	}
	cfg := ComplexConfig{}
	isNoErr(t, Parse(t.Context(), &cfg, Options{Environment: map[string]string{"FOO_STR": "foo_str"}}))
	isEqual(t, "foo_str", cfg.Foo.Str)

	type ComplexConfig2 struct {
		Foo struct {
			Bar struct {
				Str string `env:"STR"`
			} `envPrefix:"BAR_"`
			Bar2 string `env:"BAR2"`
		} `envPrefix:"FOO_"`
	}
	cfg2 := ComplexConfig2{}
	isNoErr(t, Parse(t.Context(), &cfg2, Options{Environment: map[string]string{"FOO_BAR_STR": "kek", "FOO_BAR2": "lel"}}))
	isEqual(t, "lel", cfg2.Foo.Bar2)
	isEqual(t, "kek", cfg2.Foo.Bar.Str)
}

func TestComplePrefix(t *testing.T) {
	type Config struct {
		Home string `env:"HOME"`
	}
	type ComplexConfig struct {
		Foo   Config `envPrefix:"FOO_"`
		Clean Config
		Bar   Config `envPrefix:"BAR_"`
		Blah  string `env:"BLAH"`
	}
	cfg := ComplexConfig{}
	isNoErr(t, Parse(t.Context(), &cfg, Options{
		Prefix: "T_",
		Environment: map[string]string{
			"T_FOO_HOME": "/foo",
			"T_BAR_HOME": "/bar",
			"T_BLAH":     "blahhh",
			"T_HOME":     "/clean",
		},
	}))
	isEqual(t, "/foo", cfg.Foo.Home)
	isEqual(t, "/bar", cfg.Bar.Home)
	isEqual(t, "/clean", cfg.Clean.Home)
	isEqual(t, "blahhh", cfg.Blah)
}

func TestNoEnvKey(t *testing.T) {
	type Config struct {
		Foo      string
		FooBar   string
		HTTPPort int
		bar      string
	}
	var cfg Config
	isNoErr(t, Parse(t.Context(), &cfg, Options{
		UseFieldNameByDefault: true,
		Environment: map[string]string{
			"FOO":       "fooval",
			"FOO_BAR":   "foobarval",
			"HTTP_PORT": "10",
		},
	}))
	isEqual(t, "fooval", cfg.Foo)
	isEqual(t, "foobarval", cfg.FooBar)
	isEqual(t, 10, cfg.HTTPPort)
	isEqual(t, "", cfg.bar)
}

func TestToEnv(t *testing.T) {
	for in, out := range map[string]string{
		"Foo":          "FOO",
		"FooBar":       "FOO_BAR",
		"FOOBar":       "FOO_BAR",
		"Foo____Bar":   "FOO_BAR",
		"fooBar":       "FOO_BAR",
		"Foo_Bar":      "FOO_BAR",
		"Foo__Bar":     "FOO_BAR",
		"HTTPPort":     "HTTP_PORT",
		"SSHPort":      "SSH_PORT",
		"_SSH___Port_": "SSH_PORT",
		"_PortHTTP":    "PORT_HTTP",
	} {
		t.Run(in, func(t *testing.T) {
			isEqual(t, out, toEnvName(in))
		})
	}
}

func TestErrorIs(t *testing.T) {
	err := newAggregateError(newParseError(reflect.StructField{}, nil))
	t.Run("is", func(t *testing.T) {
		isTrue(t, errors.Is(err, ParseError{}))
	})
	t.Run("is not", func(t *testing.T) {
		isFalse(t, errors.Is(err, NoParserError{}))
	})
}

type FieldParamsConfig struct {
	Simple         []string `env:"SIMPLE"`
	WithoutEnv     string
	privateWithEnv string `env:"PRIVATE_WITH_ENV"` //nolint:unused
	WithDefault    string `env:"WITH_DEFAULT" envDefault:"default"`
	Required       string `env:"REQUIRED,required"`
	File           string `env:"FILE,file"`
	Unset          string `env:"UNSET,unset"`
	NotEmpty       string `env:"NOT_EMPTY,notEmpty"`
	Expand         string `env:"EXPAND,expand"`
	NestedConfig   struct {
		Simple []string `env:"SIMPLE"`
	} `envPrefix:"NESTED_"`
}

func TestGetFieldParams(t *testing.T) {
	var config FieldParamsConfig
	params, err := GetFieldParams(&config)
	isNoErr(t, err)

	expectedParams := []FieldParams{
		{OwnKey: "SIMPLE", Key: "SIMPLE"},
		{OwnKey: "WITH_DEFAULT", Key: "WITH_DEFAULT", DefaultValue: "default", HasDefaultValue: true},
		{OwnKey: "REQUIRED", Key: "REQUIRED", Required: true},
		{OwnKey: "FILE", Key: "FILE", LoadFile: true},
		{OwnKey: "UNSET", Key: "UNSET", Unset: true},
		{OwnKey: "NOT_EMPTY", Key: "NOT_EMPTY", NotEmpty: true},
		{OwnKey: "EXPAND", Key: "EXPAND", Expand: true},
		{OwnKey: "SIMPLE", Key: "NESTED_SIMPLE"},
	}
	isTrue(t, len(params) == len(expectedParams))
	isTrue(t, areEqual(params, expectedParams))
}

func TestGetFieldParamsWithPrefix(t *testing.T) {
	var config FieldParamsConfig

	params, err := GetFieldParamsWithOptions(&config, Options{Prefix: "FOO_"})
	isNoErr(t, err)

	expectedParams := []FieldParams{
		{OwnKey: "SIMPLE", Key: "FOO_SIMPLE"},
		{OwnKey: "WITH_DEFAULT", Key: "FOO_WITH_DEFAULT", DefaultValue: "default", HasDefaultValue: true},
		{OwnKey: "REQUIRED", Key: "FOO_REQUIRED", Required: true},
		{OwnKey: "FILE", Key: "FOO_FILE", LoadFile: true},
		{OwnKey: "UNSET", Key: "FOO_UNSET", Unset: true},
		{OwnKey: "NOT_EMPTY", Key: "FOO_NOT_EMPTY", NotEmpty: true},
		{OwnKey: "EXPAND", Key: "FOO_EXPAND", Expand: true},
		{OwnKey: "SIMPLE", Key: "FOO_NESTED_SIMPLE"},
	}
	isTrue(t, len(params) == len(expectedParams))
	isTrue(t, areEqual(params, expectedParams))
}

func TestGetFieldParamsError(t *testing.T) {
	var config FieldParamsConfig

	_, err := GetFieldParams(config)
	isErrorWithMessage(t, err, "env: expected a pointer to a Struct")
	isTrue(t, errors.Is(err, NotStructPtrError{}))
}

type Conf struct {
	Foo string `env:"FOO" envDefault:"bar"`
}

func TestParseAs(t *testing.T) {
	config, err := ParseAs[Conf](t.Context(), Options{
		Environment: map[string]string{
			"FOO": "not bar",
		},
	})
	isNoErr(t, err)
	isEqual(t, "not bar", config.Foo)
}

type ConfRequired struct {
	Foo string `env:"FOO,required"`
}

func TestMust(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		defer func() {
			err := recover()
			isErrorWithMessage(t, err.(error), `env: required environment variable "FOO" is not set`)
		}()
		conf, err := ParseAs[ConfRequired](t.Context())
		isNoErr(t, err)
		isEqual(t, "", conf.Foo)
	})
	t.Run("success", func(t *testing.T) {
		t.Setenv("FOO", "bar")
		conf, err := ParseAs[ConfRequired](t.Context())
		isNoErr(t, err)
		isEqual(t, "bar", conf.Foo)
	})
}
*/

func isTrue(tb testing.TB, b bool) {
	tb.Helper()

	if !b {
		tb.Fatalf("expected true, got false")
	}
}

func isFalse(tb testing.TB, b bool) {
	tb.Helper()

	if b {
		tb.Fatalf("expected false, got true")
	}
}

func isErrorWithMessage(tb testing.TB, err error, msg string) {
	tb.Helper()

	if err == nil {
		tb.Fatalf("expected error, got nil")
	}

	if msg != err.Error() {
		tb.Fatalf("expected error message %q, got %q", msg, err.Error())
	}
}

func isNoErr(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}

func isEqual(tb testing.TB, a, b any) {
	tb.Helper()

	if areEqual(a, b) {
		return
	}

	tb.Fatalf("expected %#v (type %T) == %#v (type %T)", a, a, b, b)
}

// copied from https://github.com/matryer/is
func areEqual(a, b any) bool {
	if isNil(a) && isNil(b) {
		return true
	}
	if isNil(a) || isNil(b) {
		return false
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	return aValue == bValue
}

// copied from https://github.com/matryer/is
func isNil(object any) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

func TestParseOverride(t *testing.T) {
	t.Skip()
	type config struct {
		Interval time.Duration `env:"INTERVAL"`
	}

	var cfg config

	isNoErr(t, Parse(t.Context(), &cfg,
		// WithParserFunc(reflect.TypeFor[time.Duration](), func(_ context.Context, value string) (any, error) {
		// 	intervalI, err := strconv.Atoi(value)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	return time.Duration(intervalI), nil
		// }),
		WithEnvironment(map[string]string{
			"INTERVAL": "1",
		}),
	))
}

type Password []byte

func (p *Password) UnmarshalText(text []byte) error {
	out, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return err
	}
	*p = out
	return nil
}

type UsernameAndPassword struct {
	Username string    `env:"USER"`
	Password *Password `env:"PWD"`
}

func TestBase64Password(t *testing.T) {
	t.Setenv("USER", "admin")
	t.Setenv("PWD", base64.StdEncoding.EncodeToString([]byte("admin123")))
	var c UsernameAndPassword
	isNoErr(t, Parse(t.Context(), &c))
	isEqual(t, "admin", c.Username)
	isEqual(t, "admin123", string(*c.Password))
}

func TestIssue304(t *testing.T) {
	t.Setenv("BACKEND_URL", "https://google.com")
	type Config struct {
		BackendURL string `envDefault:"localhost:8000"`
	}
	cfg, err := ParseAs[Config](t.Context())
	isNoErr(t, err)
	isEqual(t, "https://google.com", cfg.BackendURL)
}

func TestIssue234(t *testing.T) {
	type Test struct {
		Str string `env:"TEST"`
	}
	type ComplexConfig struct {
		Foo   *Test `envPrefix:"FOO_"`
		Bar   Test  `envPrefix:"BAR_"`
		Clean *Test
	}

	t.Setenv("FOO_TEST", "kek")
	t.Setenv("BAR_TEST", "lel")

	cfg := ComplexConfig{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "kek", cfg.Foo.Str)
	isEqual(t, "lel", cfg.Bar.Str)
}

type Issue308 struct {
	Inner Issue308Map `env:"A_MAP"`
}

type Issue308Map map[string][]string

func (rc *Issue308Map) UnmarshalText(b []byte) error {
	m := map[string][]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*rc = Issue308Map(m)
	return nil
}

func TestIssue308(t *testing.T) {
	t.Setenv("A_MAP", `{"FOO":["BAR", "ZAZ"]}`)

	cfg := Issue308{}
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, Issue308Map{"FOO": []string{"BAR", "ZAZ"}}, cfg.Inner)
}

func TestIssue317(t *testing.T) {
	type TestConfig struct {
		U1 *url.URL `env:"U1"`
		U2 *url.URL `env:"U2"`
	}
	cases := []struct {
		desc                   string
		environment            map[string]string
		expectedU1, expectedU2 *url.URL
	}{
		{
			desc:        "unset",
			environment: map[string]string{},
			expectedU1:  nil,
			expectedU2:  &url.URL{},
		},
		{
			desc:        "empty",
			environment: map[string]string{"U1": "", "U2": ""},
			expectedU1:  nil,
			expectedU2:  &url.URL{},
		},
		{
			desc:        "set",
			environment: map[string]string{"U1": "https://example.com/"},
			expectedU1:  &url.URL{Scheme: "https", Host: "example.com", Path: "/"},
			expectedU2:  &url.URL{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cfg := TestConfig{}
			err := Parse(t.Context(), &cfg, WithEnvironment(tc.environment))
			isNoErr(t, err)
			isEqual(t, tc.expectedU1, cfg.U1)
			isEqual(t, tc.expectedU2, cfg.U2)
		})
	}
}

func TestIssue310(t *testing.T) {
	type TestConfig struct {
		URL *url.URL
	}
	cfg, err := ParseAs[TestConfig](t.Context())
	isNoErr(t, err)
	isEqual(t, nil, cfg.URL)
}

func TestMultipleTagOptions(t *testing.T) {
	type TestConfig struct {
		URL *url.URL `env:"URL"`
	}
	t.Run("unset", func(t *testing.T) {
		cfg, err := ParseAs[TestConfig](t.Context())
		isNoErr(t, err)
		isEqual(t, &url.URL{}, cfg.URL)
	})
	t.Run("empty", func(t *testing.T) {
		t.Setenv("URL", "")
		cfg, err := ParseAs[TestConfig](t.Context())
		isNoErr(t, err)
		isEqual(t, &url.URL{}, cfg.URL)
	})
	t.Run("set", func(t *testing.T) {
		t.Setenv("URL", "https://github.com/caarlos0")
		cfg, err := ParseAs[TestConfig](t.Context())
		isNoErr(t, err)
		isEqual(t, &url.URL{Scheme: "https", Host: "github.com", Path: "/caarlos0"}, cfg.URL)
		isEqual(t, "", os.Getenv("URL"))
	})
}

func TestIssue298(t *testing.T) {
	type Test struct {
		Str string `env:"STR"`
		Num int    `env:"NUM"`
	}
	type ComplexConfig struct {
		Foo *[]Test `envPrefix:"FOO_"`
		Bar []Test  `envPrefix:"BAR"`
		Baz []Test  `env:""`
	}

	t.Setenv("FOO_0_STR", "f0t")
	t.Setenv("FOO_0_NUM", "101")
	t.Setenv("FOO_1_STR", "f1t")
	t.Setenv("FOO_1_NUM", "111")

	t.Setenv("BAR_0_STR", "b0t")
	// t.Setenv("BAR_0_NUM", "202") // Not overridden
	t.Setenv("BAR_1_STR", "b1t")
	t.Setenv("BAR_1_NUM", "212")

	t.Setenv("0_STR", "bt")
	t.Setenv("1_NUM", "10")

	sample := make([]Test, 1)
	sample[0].Str = "overridden text"
	sample[0].Num = 99999999
	cfg := ComplexConfig{Bar: sample}

	isNoErr(t, Parse(t.Context(), &cfg))

	isEqual(t, "f0t", (*cfg.Foo)[0].Str)
	isEqual(t, 101, (*cfg.Foo)[0].Num)
	isEqual(t, "f1t", (*cfg.Foo)[1].Str)
	isEqual(t, 111, (*cfg.Foo)[1].Num)

	isEqual(t, "b0t", cfg.Bar[0].Str)
	isEqual(t, 99999999, cfg.Bar[0].Num)
	isEqual(t, "b1t", cfg.Bar[1].Str)
	isEqual(t, 212, cfg.Bar[1].Num)

	isEqual(t, "bt", cfg.Baz[0].Str)
	isEqual(t, 0, cfg.Baz[0].Num)
	isEqual(t, "", cfg.Baz[1].Str)
	isEqual(t, 10, cfg.Baz[1].Num)
}

func TestIssue298ErrorNestedFieldRequiredNotSet(t *testing.T) {
	type Test struct {
		Str string `env:"STR,required"`
		Num int    `env:"NUM"`
	}
	type ComplexConfig struct {
		Foo *[]Test `envPrefix:"FOO"`
	}

	t.Setenv("FOO_0_NUM", "101")

	cfg := ComplexConfig{}
	err := Parse(t.Context(), &cfg)
	isErrorWithMessage(t, err, `env: required environment variable "FOO_0_STR" is not set`)
	isTrue(t, errors.Is(err, EmptyVarError{}))
}

func TestIssue320(t *testing.T) {
	type Test struct {
		Str string `env:"STR"`
		Num int    `env:"NUM"`
	}
	type ComplexConfig struct {
		Foo *[]Test `envPrefix:"FOO_"`
		Bar []Test  `envPrefix:"BAR"`
		Baz []Test  `env:""`
	}

	cfg := ComplexConfig{}

	isNoErr(t, Parse(t.Context(), &cfg))

	isEqual(t, cfg.Foo, nil)
	isEqual(t, cfg.Bar, nil)
	isEqual(t, cfg.Baz, nil)
}

func TestParseRenamedDefault(t *testing.T) {
	type config struct {
		Str string `env:"STR" envDefault:"bar"`
	}

	cfg := &config{}
	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "bar", cfg.Str)

	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "foo", cfg.Str)
}

func TestSetDefaultsForZeroValuesOnly(t *testing.T) {
	type config struct {
		Str string  `env:"STR" envDefault:"foo"`
		Int int     `env:"INT" envDefault:"42"`
		URL url.URL `env:"URL" envDefault:"https://github.com/caarlos0"`
	}
	defURL, err := url.Parse("https://github.com/caarlos0")
	isNoErr(t, err)

	u, err := url.Parse("https://localhost/foo")
	isNoErr(t, err)

	for _, tc := range []struct {
		Name     string
		Options  []Option
		Expected config
	}{
		{
			Name:    "true",
			Options: []Option{},
			Expected: config{
				Str: "isSet",
				Int: 1,
				URL: *u,
			},
		},
		{
			Name:    "false",
			Options: []Option{},
			Expected: config{
				Str: "foo",
				Int: 42,
				URL: *defURL,
			},
		},
		{
			Name:    "default",
			Options: []Option{},
			Expected: config{
				Str: "foo",
				Int: 42,
				URL: *defURL,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			cfg := &config{
				Str: "isSet",
				Int: 1,
				URL: *u,
			}
			isNoErr(t, Parse(t.Context(), cfg, tc.Options...))
			isEqual(t, tc.Expected, *cfg)
		})
	}
}

func TestParseRenamedPrefix(t *testing.T) {
	type Config struct {
		Str string `env:"STR"`
	}
	type ComplexConfig struct {
		Foo Config `envPrefix:"BAR_"`
	}

	t.Setenv("FOO_STR", "101")
	t.Setenv("BAR_STR", "202")
	t.Setenv("APP_BAR_STR", "303")

	cfg := &ComplexConfig{}
	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "202", cfg.Foo.Str)

	isNoErr(t, Parse(t.Context(), cfg, WithPrefix("APP_")))
	isEqual(t, "303", cfg.Foo.Str)

	isNoErr(t, Parse(t.Context(), cfg))
	isEqual(t, "101", cfg.Foo.Str)
}

func TestFieldIgnored(t *testing.T) {
	type Test struct {
		Foo string `env:"FOO"`
		Bar string `env:"BAR,-"`
	}
	type ComplexConfig struct {
		Str string `env:"STR"`
		Foo Test   `env:"FOO" envPrefix:"FOO_"`
		Bar Test   `env:"-" envPrefix:"BAR_"`
	}
	t.Setenv("STR", "101")
	t.Setenv("FOO_FOO", "202")
	t.Setenv("FOO_BAR", "303")
	t.Setenv("BAR_FOO", "404")
	t.Setenv("BAR_BAR", "505")

	var cfg ComplexConfig
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "101", cfg.Str)
	isEqual(t, "202", cfg.Foo.Foo)
	isEqual(t, "", cfg.Foo.Bar)
	isEqual(t, "", cfg.Bar.Foo)
	isEqual(t, "", cfg.Bar.Bar)
}

func TestNoEnvKeyIgnored(t *testing.T) {
	type Config struct {
		Foo    string `env:"-"`
		FooBar string
	}

	t.Setenv("FOO", "101")
	t.Setenv("FOO_BAR", "202")

	var cfg Config
	isNoErr(t, Parse(t.Context(), &cfg))
	isEqual(t, "", cfg.Foo)
	isEqual(t, "202", cfg.FooBar)
}

func TestIssue339(t *testing.T) {
	t.Run("Should parse with bool ptr set and env undefined", func(t *testing.T) {
		existingValue := true
		cfg := Config{
			BoolPtr: &existingValue,
		}

		isNoErr(t, Parse(t.Context(), &cfg, WithEnvironment(map[string]string{
			"STRING":  tos("sdvsdf"),
			"STRINGS": tos("wetwer"),
			// "BOOL":               tos(""),
			// "BOOLS":              tos("", ""),
			"INT":                tos("0"),
			"INTS":               tos("0"),
			"INT8":               tos("0"),
			"INT8S":              tos("0"),
			"INT16":              tos("0"),
			"INT16S":             tos("0"),
			"INT32":              tos("0"),
			"INT32S":             tos("0"),
			"INT64":              tos("0"),
			"INT64S":             tos("0"),
			"UINT":               tos("0"),
			"UINTS":              tos("0"),
			"UINT8":              tos("0"),
			"UINT8S":             tos("0"),
			"UINT16":             tos("0"),
			"UINT16S":            tos("0"),
			"UINT32":             tos("0"),
			"UINT32S":            tos("0"),
			"UINT64":             tos("0"),
			"UINT64S":            tos("0"),
			"FLOAT32":            tos("0"),
			"FLOAT32S":           tos("0"),
			"FLOAT64":            tos("0"),
			"FLOAT64S":           tos("0"),
			"DURATION":           tos("0"),
			"DURATIONS":          tos("0"),
			"LOCATION":           tos(""),
			"LOCATIONS":          tos("", ""),
			"UNMARSHALER":        tos(time.Second),
			"UNMARSHALERS":       tos(time.Second, time.Minute),
			"URL":                tos("http://example.com"),
			"URLS":               tos("http://example.com", "http://example.org"),
			"SEPSTRINGS":         "a" + ":" + "b",
			"NONDEFINED_STR":     "nonDefinedStr",
			"PRF_NONDEFINED_STR": "nonDefinedStr",
			"FOO":                "bar",
		})))

		isEqual(t, &existingValue, cfg.BoolPtr)
	})

	t.Run("Should parse with bool ptr set and env defined", func(t *testing.T) {
		existingValue := true
		cfg := Config{
			BoolPtr: &existingValue,
		}

		newValue := false
		t.Setenv("BOOL", strconv.FormatBool(newValue))

		isNoErr(t, Parse(t.Context(), &cfg))

		isEqual(t, &newValue, cfg.BoolPtr)
	})

	t.Run("Should parse with string ptr set and env undefined", func(t *testing.T) {
		existingValue := "one"
		cfg := Config{
			StringPtr: &existingValue,
		}

		isNoErr(t, Parse(t.Context(), &cfg))

		isEqual(t, &existingValue, cfg.StringPtr)
	})

	t.Run("Should parse with string ptr set and env defined", func(t *testing.T) {
		existingValue := "one"
		cfg := Config{
			StringPtr: &existingValue,
		}

		newValue := "two"
		t.Setenv("STRING", newValue)

		isNoErr(t, Parse(t.Context(), &cfg))

		isEqual(t, &newValue, cfg.StringPtr)
	})
}

func TestIssue350(t *testing.T) {

	type Config struct {
		Map map[string]string `env:"MAP"`
	}

	var cfg Config
	isNoErr(t, Parse(t.Context(), &cfg, WithEnvironment(map[string]string{
		"MAP": "url:https://foo.bar:2030",
	})))
	isEqual(t, map[string]string{"url": "https://foo.bar:2030"}, cfg.Map)
}

func TestEnvBleed(t *testing.T) {
	type Config struct {
		Foo string `env:"FOO" envDefault:""`
	}

	t.Run("Default env with value", func(t *testing.T) {
		var cfg Config
		isNoErr(t, Parse(t.Context(), &cfg, WithEnvironment(map[string]string{
			"FOO": "101",
		})))
		isEqual(t, "101", cfg.Foo)
	})

	t.Run("Empty env without value", func(t *testing.T) {
		var cfg Config
		isNoErr(t, Parse(t.Context(), &cfg))
		isEqual(t, "", cfg.Foo)
	})

	t.Run("Custom env with overwritten value", func(t *testing.T) {
		var cfg Config
		isNoErr(t, Parse(t.Context(), &cfg, WithEnvironment(map[string]string{
			"FOO": "202",
		})))
		isEqual(t, "202", cfg.Foo)
	})

	t.Run("Custom env without value", func(t *testing.T) {
		var cfg Config
		isNoErr(t, Parse(t.Context(), &cfg, WithEnvironment(map[string]string{
			"BAR": "202",
		})))
		isEqual(t, "", cfg.Foo)
	})
}
