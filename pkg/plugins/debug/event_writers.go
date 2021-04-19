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

package debug

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/xmidt-org/ears/pkg/event"
)

func (s *SendStderr) Write(e event.Event) error {
	return write(os.Stderr, e)
}

func (s *SendStdout) Write(e event.Event) error {
	return write(os.Stdout, e)
}

func (s *SendSlice) Write(e event.Event) error {
	s.Lock()
	defer s.Unlock()
	if s.events == nil {
		s.events = []event.Event{}
	}
	s.events = append(s.events, e)
	return nil
}

func (s *SendSlice) Events() []event.Event {
	return s.events
}

func write(w io.Writer, e event.Event) error {
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	out := "<nil>"
	if e != nil && e.Payload() != nil {
		buf, err := json.Marshal(e.Payload())
		if err != nil {
			return err
		}
		out = fmt.Sprint(string(buf))
	}
	fmt.Fprintln(w, out)
	return nil
}
