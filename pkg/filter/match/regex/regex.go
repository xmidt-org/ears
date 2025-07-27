// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package regex

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/xmidt-org/ears/pkg/event"
)

type Matcher struct {
	r    *regexp.Regexp
	path string
}

// TODO: Possibly add POSIX option to pattern
func NewMatcher(pattern interface{}, path string) (*Matcher, error) {
	p, ok := pattern.(string)
	if !ok {
		pp, ok := pattern.(*string)
		if !ok {
			return nil, errors.New("regex pattern is not a string")
		}
		p = *pp
	}
	r, err := regexp.Compile(p)
	if err != nil {
		return nil, err
	}

	return &Matcher{r: r, path: path}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || m.r == nil || event == nil {
		return false
	}
	obj, _, _ := event.GetPathValue(m.path)
	eventString := ""
	switch obj := obj.(type) {
	case string:
		eventString = obj
	case []byte:
		eventString = string(obj)
	case nil:
		return false
	default:
		buf, err := json.Marshal(obj)
		if err != nil {
			return false
		}
		eventString = string(buf)
	}
	return m.r.MatchString(eventString)
}
