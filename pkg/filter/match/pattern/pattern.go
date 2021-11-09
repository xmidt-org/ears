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
	exactArrayMatch bool
	matchMetadata   bool
}

func NewMatcher(p interface{}, exactArrayMatch bool, matchMetadata bool) (*Matcher, error) {
	return &Matcher{pattern: p, exactArrayMatch: exactArrayMatch, matchMetadata: matchMetadata}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || m.pattern == nil || event == nil {
		return false
	}
	if m.matchMetadata {
		return m.contains(event.Metadata(), m.pattern)
	} else {
		return m.contains(event.Payload(), m.pattern)
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
