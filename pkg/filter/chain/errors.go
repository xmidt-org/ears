package chain

import "fmt"

func (e *InvalidArgumentError) Unwrap() error {
	return e.Err
}

func (e *InvalidArgumentError) Error() string {
	return errToString("InvalidArgumentError", e.Err)
}

func errToString(name string, err error) string {
	if err == nil {
		return name
	}

	return fmt.Sprintf("%s: %s", name, err)
}
