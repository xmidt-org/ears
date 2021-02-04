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
	"bytes"
)

type text struct {
	base
	content string
}

// ==============================================================
// Interface Processor
// ==============================================================

func (t *text) Parse(content string) error {
	t.init()

	t.RLock()
	tpl, err := t.tpl.Parse(content)
	t.RUnlock()
	if err != nil {
		return &ParseError{
			Err:     err,
			Content: content,
		}
	}

	t.Lock()
	t.content = content
	t.tpl = tpl
	t.Unlock()

	return nil
}

func (t *text) Apply(values interface{}) (string, error) {
	t.init()

	var res bytes.Buffer

	t.RLock()
	defer t.RUnlock()

	if err := t.tpl.Execute(&res, values); err != nil {
		if isMissingKeyError(err) {
			return "", &MissingKeyError{
				Err:     err,
				Key:     getMissingKeyFromError(err),
				Content: t.content,
				Values:  values,
			}

		}

		return "", &ApplyValueError{
			Err:     err,
			Content: t.content,
			Values:  values,
		}
	}

	return res.String(), nil
}
