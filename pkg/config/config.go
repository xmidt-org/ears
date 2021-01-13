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
)

func NewConfig(data interface{}, target interface{}) error {
	// fmt.Println("DATA (", reflect.TypeOf(data), "):", data)
	// fmt.Println("TARGET (", reflect.TypeOf(target), "):", target)

	if target == nil {
		return fmt.Errorf("target is nil")
	}

	if data == nil {
		return fmt.Errorf("data is nil")
	}

	switch data.(type) {
	case string, []byte, []rune:
		// fmt.Println("  FROM YAML ---")
		return FromYAML(data, target)

	default:
		// fmt.Println("  default ---")
		dt := strings.TrimPrefix(reflect.TypeOf(data).String(), "*")
		tt := strings.TrimPrefix(reflect.TypeOf(target).String(), "*")

		if dt == tt {
			dv := reflect.ValueOf(data)
			for dv.Kind() == reflect.Ptr || dv.Kind() == reflect.Interface {
				dv = dv.Elem()
			}

			tv := reflect.ValueOf(target)
			for tv.Kind() == reflect.Ptr || tv.Kind() == reflect.Interface {
				tv = tv.Elem()
			}
			tv.Set(dv)

			return nil
		}

	}

	return &Error{
		Err: fmt.Errorf("could not set config object"),
	}
}
