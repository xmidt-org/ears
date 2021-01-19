package validation_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/xmidt-org/ears/pkg/validation"
)

func TestErrorMessage(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "Error", err: &validation.Error{}},

		{
			name: "Error_Err",
			err:  &validation.Error{Err: fmt.Errorf("wrapped error")},
		},
		{name: "Errors", err: &validation.Errors{}},

		{
			name: "Errors_Err",
			err: &validation.Errors{Errs: []error{
				fmt.Errorf("wrapped 1 error"),
				fmt.Errorf("wrapped 2 error"),
				fmt.Errorf("wrapped 3 error"),
			}},
		},

		{name: "ProcessingError", err: &validation.ProcessingError{}},

		{
			name: "ProcessingError_Err",
			err:  &validation.ProcessingError{Err: fmt.Errorf("wrapped error")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := goldie.New(t, goldie.WithTestNameForDir(true))
			g.Assert(t, tc.name, []byte(fmt.Sprint(tc.err)))
			g.Assert(t, tc.name+"_unwrapped", []byte(fmt.Sprint(errors.Unwrap(tc.err))))

		})
	}

}
