package errors

import (
	"fmt"
	"sort"
	"strings"
)

/*
String will generate an error string that conforms to our
error coding style.  This will:

  * Print the name of the error struct
  * Enclose any key sorted values within parenthesis
  * Append any wrapped errors to the end of the message (separated by a ":")

*/
func String(name string, values map[string]interface{}, wrapped error) string {
	msgs := []string{}

	if name != "" {
		msgs = append(msgs, name)
	}

	if len(values) > 0 {
		keys := []string{}
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		kv := []string{}
		for _, k := range keys {
			kv = append(kv, fmt.Sprintf("%s=%v", k, values[k]))
		}

		msgs = append(msgs, "("+strings.Join(kv, " ")+")")

	}

	msg := strings.Join(msgs, " ")
	errMsg := ""
	if wrapped != nil {
		errMsg = wrapped.Error()
	}

	switch {
	case msg == "" && errMsg == "":
		return "unspecified error"
	case msg == "" && errMsg != "":
		return errMsg
	case msg != "" && errMsg == "":
		return msg
	}

	return fmt.Sprintf("%s: %s", msg, errMsg)
}
