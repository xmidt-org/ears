// Copyright 2021 Comcast Cable Communications Management, LLC
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

package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xmidt-org/ears/pkg/errs"
)

// NewConfig will create process the data and will fill the
// target with the values supplied.  For the data object,
// this function supports YAML, JSON, the target type,
// and a pointer to the target type.
func NewConfig(data interface{}, target interface{}) error {

	if target == nil {
		return &InvalidArgumentError{Err: fmt.Errorf("target is nil")}
	}

	if data == nil {
		return &InvalidArgumentError{Err: fmt.Errorf("data is nil")}
	}

	switch d := data.(type) {
	case []rune:
		err := FromYAML(string(d), target)
		if err != nil {
			return &DataParseError{Err: err}
		}
		return nil
	case string, []byte:
		err := FromYAML(d, target)
		if err != nil {
			return &DataParseError{Err: err}
		}
		return nil

	default:
		dt := strings.TrimPrefix(reflect.TypeOf(d).String(), "*")
		tt := strings.TrimPrefix(reflect.TypeOf(target).String(), "*")

		if dt == tt {
			dv := reflect.ValueOf(d)
			for dv.Kind() == reflect.Ptr || dv.Kind() == reflect.Interface {
				dv = dv.Elem()
			}

			tv := reflect.ValueOf(target)
			for tv.Kind() == reflect.Ptr || tv.Kind() == reflect.Interface {
				tv = tv.Elem()
			}

			if !tv.CanSet() {
				return &InvalidArgumentError{Err: fmt.Errorf("target is not a pointer")}
			}
			tv.Set(dv)

			return nil
		} else {
			return &InvalidArgumentError{
				Err: fmt.Errorf(errs.String(
					"incompatible object types",
					map[string]interface{}{
						"data":   dt,
						"config": tt,
					},
					nil,
				)),
			}
		}
	}
}
