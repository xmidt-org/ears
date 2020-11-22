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

// a TTLFilter filters an event when it is too old
type TTLFilter struct {
	TTL    int    // ttl in seconds
	TsPath string // path to timestamp information in payload
}

// Filter filters event if it hhas expired
func (mf *TTLFilter) Filter(ctx context.Context, event Event) ([]Event, error) {
	// never filters
	events := make([]Event, 0)
	//TODO: implement filter logic
	events = append(events, event)
	return events, nil
}
