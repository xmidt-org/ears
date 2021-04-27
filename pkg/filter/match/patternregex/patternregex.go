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

package patternregex

import (
	"github.com/xmidt-org/ears/pkg/event"
	"regexp"
)

type Matcher struct {
	p interface{}
}

func NewMatcher(p interface{}) (*Matcher, error) {
	return &Matcher{p: p}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || m.p == nil || event == nil {
		return false
	}
	return contains(event.Payload(), m.p)
}

// contains is a helper function to check if b is contained in a (if b is a partial of a)
func contains(a interface{}, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch b.(type) {
	case string:
		if a == nil {
			return false
		}
		_, aIsStr := a.(string)
		if !aIsStr {
			return false
		}
		// interpret b as regex
		r, err := regexp.Compile(b.(string))
		if err != nil {
			return false
		}
		if !r.MatchString(a.(string)) {
			return false
		}
		return true
	}
	switch b.(type) {
	case map[string]interface{}:
		switch a.(type) {
		case map[string]interface{}:
			for k, vb := range b.(map[string]interface{}) {
				c := contains(a.(map[string]interface{})[k], vb)
				if !c {
					return false
				}
			}
		default:
			return false
		}
	case []interface{}:
		switch a.(type) {
		case []interface{}:
			// check if all fields in b are in a (in any order)
			for _, vb := range b.([]interface{}) {
				present := false
				for _, va := range a.([]interface{}) {
					// this supports partial matches in deeply structured array elements
					c := contains(va, vb)
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
