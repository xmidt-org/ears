package app

import (
	"errors"
	"testing"
)

type MyTestError struct {
}

func (e *MyTestError) Error() string {
	return "A unit test error"
}

func TestErrors(t *testing.T) {

	err := &Error{
		Err: &MyTestError{},
	}
	var myTestErr *MyTestError
	if !errors.As(err, &myTestErr) {
		t.Errorf("Error type check failed")
	}

	err2 := &Error{
		Err: &MyTestError{},
	}
	if !errors.Is(err, err2) {
		t.Error("Error comparison failed")
	}

	err3 := &InvalidOptionError{
		Err: &MyTestError{},
	}
	if errors.Is(err, err3) {
		t.Error("Error comparison should fail but not failing")
	}

	if !errors.As(err3, &myTestErr) {
		t.Errorf("Error type check failed")
	}

	err4 := &InvalidOptionError{
		Err: &MyTestError{},
	}
	if !errors.Is(err3, err4) {
		t.Error("Error comparison failed")
	}

}
