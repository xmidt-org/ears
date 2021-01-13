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

package match

import (
	"context"
	"regexp"

	"github.com/xmidt-org/ears/pkg/event"
)

type Matcher struct {
	r *regexp.Regexp
}

// TODO: Possibly add POSIX option to pattern
func NewMatcher(pattern string) (*Matcher, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Matcher{r: r}, nil
}

func (m *Matcher) Match(ctx context.Context, event event.Event) bool {
	if m == nil || m.r == nil || event == nil {
		return false
	}

	p, ok := event.Payload().(string)
	if !ok {
		return false
	}

	return m.r.MatchString(p)
}
