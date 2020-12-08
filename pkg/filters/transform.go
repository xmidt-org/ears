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

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
)

type TransformFilter struct {
	Transformation interface{} // transformation instructions for event payload
}

// Filter transforms an event according to its transformation spec
func (mf *TransformFilter) Filter(ctx context.Context, evt event.Event) ([]event.Event, error) {
	// identity transform
	events := make([]event.Event, 0)
	//TODO: implement filter logic
	events = append(events, evt)
	return events, nil
}
