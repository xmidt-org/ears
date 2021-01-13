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
	"encoding/json"
	"fmt"

	"github.com/goccy/go-yaml"
)

func ToYAML(c interface{}) (string, error) {
	out, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func FromYAML(data interface{}, target interface{}) error {

	// Use fmt.Sprint as a best effort to convert the
	// data input into a string.
	return yaml.Unmarshal(
		[]byte(fmt.Sprintf("%s", data)),
		target,
	)

}

func ToJSON(c interface{}) (string, error) {
	out, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func FromJSON(data interface{}, target interface{}) error {

	// Use fmt.Sprint as a best effort to convert the
	// data input into a string.
	return json.Unmarshal(
		[]byte(fmt.Sprintf("%s", data)),
		target,
	)

}
