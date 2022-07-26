package errors

import (
	"fmt"
	"reflect"
)

func Wrap(cause, outer error) error {
	if cause != nil {
		reflect.ValueOf(outer).Elem().FieldByName("Cause").Set(reflect.ValueOf(cause))
		return outer
	}
	return nil
}

func Wrapf(cause error, outer error, format string, a ...interface{}) error {
	if cause != nil {
		reflect.ValueOf(outer).Elem().FieldByName("Cause").Set(reflect.ValueOf(cause))
		reflect.ValueOf(outer).Elem().FieldByName("ErrMsg").SetString(fmt.Sprintf(format, a...))
		return outer
	}
	return nil
}

func Errorf(err error, format string, a ...interface{}) error {
	reflect.ValueOf(err).Elem().FieldByName("ErrMsg").SetString(fmt.Sprintf(format, a...))
	return err
}
