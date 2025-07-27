// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
