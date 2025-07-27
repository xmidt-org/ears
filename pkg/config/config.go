// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
