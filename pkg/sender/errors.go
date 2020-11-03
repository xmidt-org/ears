package sender

import "fmt"

func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

func (e *InvalidConfigError) Error() string {
	return errToString("InvalidConfigError", e.Err)
}

func errToString(name string, err error) string {
	if err == nil {
		return fmt.Sprintf("%s: unknown", name)
	}

	return fmt.Sprintf("%s: %s", name, err)
}
