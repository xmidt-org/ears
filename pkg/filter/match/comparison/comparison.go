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

package comparison

import (
	"github.com/xmidt-org/ears/pkg/event"
	"reflect"
	"strings"
)

type Matcher struct {
	comparison    *Comparison
	patternsLogic string
}

// so far only Equals is supported but in combination with deny this can also be used as NotEqual
type Comparison struct {
	Equal    []map[string]interface{} `json:"equal,omitempty"`
	NotEqual []map[string]interface{} `json:"notEqual,omitempty"`
	//GreaterThan []map[string]interface{} `json:"greaterThan,omitempty"`
	//LessThan    []map[string]interface{} `json:"lessThan,omitempty"`
}

func NewMatcher(comparison *Comparison, patternsLogic string) (*Matcher, error) {
	return &Matcher{comparison: comparison, patternsLogic: patternsLogic}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || m.comparison == nil || event == nil {
		return false
	}
	return m.compare(event, m.comparison)
}

func (m *Matcher) compare(evt event.Event, cmp *Comparison) bool {
	if evt == nil || cmp == nil {
		return true
	}
	andLogic := true
	if strings.ToLower(m.patternsLogic) == "or" {
		andLogic = false
	}
	foundOneEqual := false
	for _, eq := range cmp.Equal {
		for b, a := range eq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if reflect.DeepEqual(aObj, bObj) {
				foundOneEqual = true
				if !andLogic {
					return true
				}
			} else {
				if andLogic {
					return false
				}
			}
		}
	}
	for _, eq := range cmp.NotEqual {
		for b, a := range eq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if !reflect.DeepEqual(aObj, bObj) {
				foundOneEqual = true
				if !andLogic {
					return true
				}
			} else {
				if andLogic {
					return false
				}
			}
		}
	}
	if foundOneEqual || (len(cmp.Equal) == 0 && len(cmp.NotEqual) == 0) {
		return true
	}
	return false
}
