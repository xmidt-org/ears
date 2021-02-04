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

	//	encoder "encoding/json"
	"errors"
	"fmt"

	// encoder "github.com/goccy/go-yaml"
	encoder "github.com/ghodss/yaml" // Keeps the `|` when writing out docs
)

type yaml struct {
	base
	content string
}

func (y *yaml) init() {
	y.base.init()
}

func (y *yaml) Parse(content string) error {
	y.init()

	y.Lock()
	defer y.Unlock()

	var tree interface{}
	err := encoder.Unmarshal([]byte(content), &tree)

	if err != nil {
		return &ParseError{
			Err:     err,
			Content: content,
		}
	}

	// Even though the parser doesn't return an error, the object
	// returned can be a plain string.  This means the file really
	// wasn't yaml and we fail here as well.
	if _, ok := tree.(string); ok {
		// TODO: Return a proper error object
		return &ParseError{
			Err:     errors.New("Content is a plain text document"),
			Content: content,
		}
	}

	// Store the content.  This is the easiest way to "deep copy"
	// although we pay the yaml parse penalty
	y.content = content
	return nil
}

func (y *yaml) Apply(values interface{}) (string, error) {
	y.init()

	if y.content == "" {
		// TODO: Create proper error
		return "", &ApplyValueError{
			Err:     errors.New("Parse needs to be called first"),
			Content: y.content,
			Values:  values,
		}
	}

	var tree interface{}

	err := encoder.Unmarshal([]byte(y.content), &tree)
	if err != nil {
		return "", &ParseError{
			Err:     err,
			Content: y.content,
		}
	}

	// Even though the parser doesn't return an error, the object
	// returned can be a plain string.  This means the file really
	// wasn't yaml and we fail here as well.
	if _, ok := tree.(string); ok {
		// TODO: Return a proper error object
		return "", &ParseError{
			Err:     errors.New("Content is a plain text document"),
			Content: y.content,
		}
	}

	err = walk(y.tpl, tree, values)
	if err != nil {
		// Walking will return a proper error.  Just annotate
		return "", fmt.Errorf("Could not Apply values: %w", err)
	}

	content, err := encoder.Marshal(tree)
	if err != nil {
		return "", &ApplyValueError{
			Err:     fmt.Errorf("Could not marshal yaml: %w", err),
			Content: y.content,
			Values:  values,
		}
	}

	return string(content), nil
}
