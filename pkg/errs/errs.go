// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errs

import (
	"fmt"
	"reflect"
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
func String(msg string, values map[string]interface{}, wrapped error) string {
	msgs := []string{}

	if msg != "" {
		msgs = append(msgs, msg)
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

	msg = strings.Join(msgs, " ")
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

func Type(err error) string {
	if err == nil {
		return "<nil>"
	}

	return reflect.TypeOf(err).String()
}
