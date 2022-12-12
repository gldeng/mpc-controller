package errors

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

type ErrTyp struct {
	Msg string
	Pre error
}

func (e *ErrTyp) Error() string {
	return e.Msg
}

func TestWrap(t *testing.T) {
	outer := &ErrTyp{}

	outerErr := Wrap(nil, outer)
	require.Nil(t, outerErr)

	cause := errors.New("oops!")
	outerErr = Wrap(cause, outer)
	outer = outerErr.(*ErrTyp)
	require.Equal(t, cause, outer.Pre)
	require.Equal(t, "oops!", outer.Pre.Error())
}

func TestWrapf(t *testing.T) {
	outer := &ErrTyp{}

	status := 1
	outerErr := Wrapf(nil, outer, "something wrong, status: %v", status)
	require.Nil(t, outerErr)

	cause := errors.New("oops!")
	outerErr = Wrapf(cause, outer, "something wrong, status: %v", status)
	outer = outerErr.(*ErrTyp)
	require.Equal(t, cause, outer.Pre)
	require.Equal(t, "oops!", outer.Pre.Error())
	require.Equal(t, "something wrong, status: 1", outer.Msg)
}

func TestErrorf(t *testing.T) {
	err := &ErrTyp{}
	status := 1
	errCreated := Errorf(err, "something wrong, status: %v", status)
	require.Equal(t, "something wrong, status: 1", errCreated.Error())
}
