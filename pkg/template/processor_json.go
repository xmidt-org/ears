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

package template

import (
	"fmt"

	encoder "encoding/json"

	"errors"
)

type json struct {
	base
	content string
}

func (j *json) init() {
	j.base.init()

	// j.Lock()
	// defer j.Unlock()
}

func (j *json) Parse(content string) error {
	j.init()

	j.Lock()
	defer j.Unlock()

	var tree interface{}
	err := encoder.Unmarshal([]byte(content), &tree)
	if err != nil {
		return &ParseError{
			Err:     err,
			Content: content,
		}
	}

	// Store the content.  This is the easiest way to "deep copy"
	// although we pay the yaml parse penalty
	j.content = content
	return nil
}

func (j *json) Apply(values interface{}) (string, error) {
	j.init()

	if j.content == "" {
		return "", &ApplyValueError{
			Err:     errors.New("Parse needs to be called first"),
			Content: j.content,
			Values:  values,
		}
	}

	var tree interface{}

	err := encoder.Unmarshal([]byte(j.content), &tree)

	if err != nil {
		return "", &ParseError{
			Err:     err,
			Content: j.content,
		}
	}

	err = walk(j.tpl, tree, values)
	if err != nil {
		// Walking will return a proper error.  Just annotate
		return "", fmt.Errorf("Could not Apply values: %w", err)
	}

	content, err := encoder.MarshalIndent(tree, "", "\t")

	if err != nil {
		return "", &ApplyValueError{
			Err:     fmt.Errorf("Could not marshal json: %w", err),
			Content: j.content,
			Values:  values,
		}
	}

	return string(content), err
}
