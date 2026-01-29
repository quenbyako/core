package internal

import (
	"fmt"
	"reflect"
)

type UnmarshalFuncError struct {
	Type reflect.Type
	Err  error
}

func ErrUnmarshalFunc(typ reflect.Type, err error) *UnmarshalFuncError {
	return &UnmarshalFuncError{
		Type: typ,
		Err:  err,
	}
}

func (e *UnmarshalFuncError) Error() string {
	return fmt.Sprintf("unmarshalling %v: %v", e.Type.String(), e.Err)
}
