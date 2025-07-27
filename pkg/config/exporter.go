// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
