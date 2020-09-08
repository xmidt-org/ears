package util

import (
	"encoding/json"
	"fmt"
	"time"
)

type LogListener struct {
	leftBracketCount int
	jsonStr          string
	listener         chan string
}

func NewLogListener() *LogListener {
	return &LogListener{
		leftBracketCount: 0,
		jsonStr:          "",
		listener:         make(chan string, 100),
	}
}

func (jw *LogListener) Write(p []byte) (n int, err error) {
	jw.listener <- string(p)
	return len(p), nil
}

func (jw *LogListener) Close() error {
	if jw.listener != nil {
		close(jw.listener)
	}
	return nil
}

//listen to a JSON stream until a key/value pair is found
//this function is not thread safe and is only intended for one listener at one time.
func (jw *LogListener) Listen(key string, value interface{}, timeout time.Duration) error {
	for {
		select {
		case jsonStr, ok := <-jw.listener:
			if !ok {
				break
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
	return fmt.Errorf("%s:%s not found", key, value)
}
