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

package pattern

import (
	"github.com/xmidt-org/ears/pkg/event"
)

type Matcher struct {
	pattern         interface{}
	patterns        []interface{}
	patternsLogic   string
	exactArrayMatch bool
	path            string
}

func NewMatcher(pattern interface{}, patterns []interface{}, patternsLogic string, exactArrayMatch bool, path string) (*Matcher, error) {
	return &Matcher{pattern: pattern, patterns: patterns, patternsLogic: patternsLogic, exactArrayMatch: exactArrayMatch, path: path}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || m.pattern == nil || event == nil {
		return false
	}
	obj, _, _ := event.GetPathValue(m.path)
	if len(m.patterns) > 0 {
		andLogic := true
		if m.patternsLogic == "or" || m.patternsLogic == "OR" {
			andLogic = false
		}
		found := false
		for _, pattern := range m.patterns {
			found = m.contains(obj, pattern)
			if andLogic && !found {
				return false
			}
			if !andLogic && found {
				return true
			}
		}
		if found {
			return true
		} else {
			return false
		}
	} else {
		return m.contains(obj, m.pattern)
	}
}

// contains is a helper function to check if b is contained in a (if b is a partial of a)
func (m *Matcher) contains(a interface{}, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch b := b.(type) {
	case string:
		if b == "*" && a != nil {
			return true
		}
	}
	switch b := b.(type) {
	case map[string]interface{}:
		switch a := a.(type) {
		case map[string]interface{}:
			for k, vb := range b {
				c := m.contains(a[k], vb)
				if !c {
					return false
				}
			}
		default:
			return false
		}
	case []interface{}:
		switch a := a.(type) {
		case []interface{}:
			// enforce exact array match
			if m.exactArrayMatch {
				if len(a) != len(b) {
					return false
				}
			}
			// check if all fields in b are in a (in any order)
			for _, vb := range b {
				present := false
				for _, va := range a {
					// this supports partial matches in deeply structured array elements
					c := m.contains(va, vb)
					if c {
						present = true
						break
					}
				}
				if !present {
					return false
				}
			}
		default:
			return false
		}
	default:
		return a == b
	}
	return true
}
