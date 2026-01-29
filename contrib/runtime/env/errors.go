package env

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNotStructPtr = errors.New("expected a pointer to a Struct")
	ErrValueNotSet  = errors.New("required environment variable is not set")
)

type InvalidMapItemFormatError struct {
	Item        string
	KVSeparator string
}

var _ error = (*InvalidMapItemFormatError)(nil)

func errInvalidMapItemFormat(item, kvSeparator string) error {
	return &InvalidMapItemFormatError{
		Item:        item,
		KVSeparator: kvSeparator,
	}
}

func (e *InvalidMapItemFormatError) Error() string {
	return fmt.Sprintf("invalid map item format %q, should be key%qvalue", e.Item, e.KVSeparator)
}

// FieldError occurs when it's impossible to convert the value for given type.
type FieldError struct {
	Key  string
	Type reflect.Type
	Err  error
}

var _ error = (*FieldError)(nil)

func errField(key string, typ reflect.Type, err error) *FieldError {
	return &FieldError{
		Key:  key,
		Type: typ,
		Err:  err,
	}
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("%q: %v", e.Key, e.Err)
}

func (e *FieldError) Unwrap() error { return e.Err }

// NoParserError occurs when there is no parser provided for given type.
type NoParserError struct {
	Name string
	Type reflect.Type
}

func newNoParserError(sf reflect.StructField) error {
	return NoParserError{sf.Name, sf.Type}
}

func (e NoParserError) Error() string {
	return fmt.Sprintf("no parser found for field %q of type %q", e.Name, e.Type)
}

// NoSupportedTagOptionError occurs when the given tag is not supported.
// Built-in supported tags: "", "file", "required", "unset", "notEmpty",
// "expand", "envDefault", and "envSeparator".
type NoSupportedTagOptionError struct {
	Tag string
}

func newNoSupportedTagOptionError(tag string) error {
	return NoSupportedTagOptionError{tag}
}

func (e NoSupportedTagOptionError) Error() string {
	return fmt.Sprintf("tag option %q not supported", e.Tag)
}

// EmptyVarError occurs when the variable which must be not empty is existing but has an empty value
type EmptyVarError struct {
	Key string
}

func newEmptyVarError(key string) error {
	return EmptyVarError{key}
}

func (e EmptyVarError) Error() string {
	return fmt.Sprintf("environment variable %q should not be empty", e.Key)
}

// LoadFileContentError occurs when it's impossible to load the value from the file.
type LoadFileContentError struct {
	Filename string
	Key      string
	Err      error
}

func newLoadFileContentError(filename, key string, err error) error {
	return LoadFileContentError{filename, key, err}
}

func (e LoadFileContentError) Error() string {
	return fmt.Sprintf("could not load content of file %q from variable %s: %v", e.Filename, e.Key, e.Err)
}

// ParseValueError occurs when it's impossible to convert value using given parser.
type ParseValueError struct {
	Msg string
	Err error
}

func errParseValue(message string, err error) error {
	return ParseValueError{
		Msg: message,
		Err: err,
	}
}

func (e ParseValueError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}
