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
	p interface{}
}

func NewMatcher(p interface{}) (*Matcher, error) {
	return &Matcher{p: p}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || m.p == nil || event == nil {
		return false
	}
	matches, _ := contains(event.Payload(), m.p, 0)
	return matches
}

// contains is a helper function to check if b is contained in a (if b is a partial of a)
func contains(a interface{}, b interface{}, strength int) (bool, int) {
	if a == nil && b == nil {
		return true, strength
	}
	if a == nil || b == nil {
		return false, strength
	}
	switch b.(type) {
	case string:
		if b.(string) == "*" && a != nil {
			return true, strength
		}
	}
	switch b.(type) {
	case map[string]interface{}:
		switch a.(type) {
		case map[string]interface{}:
			for k, vb := range b.(map[string]interface{}) {
				c, s := contains(a.(map[string]interface{})[k], vb, 0)
				if !c {
					return false, strength
				}
				strength += s
			}
		default:
			return false, strength
		}
	case []interface{}:
		switch a.(type) {
		case []interface{}:
			// check if all fields in b are in a (in any order)
			for _, vb := range b.([]interface{}) {
				present := false
				for _, va := range a.([]interface{}) {
					// this supports partial matches in deeply structured array elements even with wild cards etc.
					c, s := contains(va, vb, 0)
					if c {
						strength += s
						present = true
						break
					}
				}
				if !present {
					return false, strength
				}
			}
		default:
			return false, strength
		}
	default:
		equal := a == b
		if equal {
			strength++
		}
		return equal, strength
	}
	return true, strength
}
