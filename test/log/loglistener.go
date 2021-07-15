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

package log

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

type LogListener struct {
	sync.Mutex
	lastLogLine string
	listener    chan string
}

func NewLogListener() *LogListener {
	return &LogListener{
		listener: make(chan string, 100),
	}
}

func (l *LogListener) Write(p []byte) (n int, err error) {
	l.Lock()
	defer l.Unlock()
	l.lastLogLine = string(p)
	fmt.Println(l.lastLogLine)
	l.listener <- l.lastLogLine
	return len(p), nil
}

func (l *LogListener) Close() error {
	if l.listener != nil {
		close(l.listener)
	}
	return nil
}

//Assert if the last logline has the key value pair
func (l *LogListener) AssertLastLogLine(t *testing.T, key string, value interface{}) {
	var log map[string]interface{}
	err := json.Unmarshal([]byte(l.lastLogLine), &log)
	if err != nil {
		t.Errorf("Fail to unmarshal lastLogLine %s", l.lastLogLine)
		return
	}
	if log[key] != value {
		t.Errorf("log key [%s] value [%v] does not match the expected value [%v]", key, log[key], value)
	}
}

//listen to a JSON stream until a key/value pair is found
//this function is not thread safe and is only intended for one listener at a time.
func (l *LogListener) Listen(key string, value interface{}, timeout time.Duration) error {
	for {
		select {
		case jsonStr, ok := <-l.listener:
			if !ok {
				return fmt.Errorf("%s:%s not found", key, value)
			}
			var log map[string]interface{}
			err := json.Unmarshal([]byte(jsonStr), &log)
			if err != nil {
				return err
			}
			if log[key] == value {
				return nil
			}
		case <-time.After(timeout):
			return fmt.Errorf("listen timed out")
		}
	}
}
