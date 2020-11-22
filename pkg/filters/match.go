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

package filters

import "context"

const (
	MatchModeAllow int = iota
	MatchModeDeny
)

type MatchMode int

type MatchFilter struct {
	Pattern interface{} // match pattern
	Matcher Matcher
	Mode    MatchMode
}

func (mf *MatchFilter) Filter(ctx context.Context, event Event) ([]Event, error) {
	// passes if event matches
	events := []Event{}
	pass := mf.Matcher.Match(ctx, event, mf.Pattern)

	if mf.Mode == MatchModeDeny {
		pass = !pass
	}

	if pass {
		return []Event{e}, nil
	}

	return []Event{}, nil
}
