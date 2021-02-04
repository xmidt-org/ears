// Copyright 2021 Comcast Cable Communications Management, LLC
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

package ledger

import (
	"container/ring"
	"sync"
)

type Recorder interface {
	// Add adds an item to the history
	Add(e interface{}) error

	// Count returns the total number of items that have been added
	Count() int

	// Records returns the recorded items
	Records() ([]interface{}, error)
}

type LimitedRecorder interface {
	Recorder

	// Capacity returns the total capacity of the recorder
	Capacity() int
}

var _ LimitedRecorder = (*History)(nil)

type History struct {
	sync.Mutex

	size  int
	count int
	ring  *ring.Ring
}

type NotInitializedError struct{}
