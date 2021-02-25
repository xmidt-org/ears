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

package event_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/xmidt-org/ears/internal/pkg/ack"
	"github.com/xmidt-org/ears/pkg/event"
	"reflect"
	"testing"
	"time"
)

func TestEventBasic(t *testing.T) {
	ctx := context.Background()

	payload := map[string]interface{}{
		"field1": "abcd",
		"field2": 1234,
	}
	e, err := event.New(ctx, payload)
	if err != nil {
		t.Errorf("Fail to create new event %s\n", err.Error())
	}

	if e.Context() != ctx {
		t.Errorf("Fail to get context")
	}
	if !reflect.DeepEqual(e.Payload(), payload) {
		t.Errorf("Fail to match payload +%v +%v\n", e.Payload(), payload)
	}

	ctx2, _ := context.WithCancel(ctx)
	payload2 := map[string]interface{}{
		"field3": "efgh",
		"field4": 5678,
	}
	e.SetContext(ctx2)
	e.SetPayload(payload2)

	if e.Context() != ctx2 {
		t.Errorf("Fail to get context2")
	}
	if !reflect.DeepEqual(e.Payload(), payload2) {
		t.Errorf("Fail to match payload +%v +%v\n", e.Payload(), payload2)
	}
}

func TestCloneEvent(t *testing.T) {
	ctx := context.Background()

	payload := map[string]interface{}{
		"field1": "abcd",
		"field2": 1234,
		"field3": map[string]interface{}{
			"field4": 1.02,
		},
	}

	e1, err := event.New(ctx, payload)
	if err != nil {
		t.Errorf("Fail to create new event %s\n", err.Error())
	}

	e2, err := e1.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone new event %s\n", err.Error())
	}

	payload2, ok := e2.Payload().(map[string]interface{})
	if !ok {
		t.Error("Fail to cast payload to expected type")
	}
	if !reflect.DeepEqual(payload, payload2) {
		t.Errorf("payloads do not match +%v\n", payload2)
	}

	//validate that deep copy worked and that updating a payload does not
	//affect the other payload
	payload2["field1"] = "efgh"
	if payload["field1"] != "abcd" {
		t.Errorf("unexpected field1 value in payload +%v\n", payload["field1"])
	}
	payload["field2"] = 5678
	if payload2["field2"] != 1234 {
		t.Errorf("unexpected field2 value in payload2 +%v\n", payload2["field2"])
	}
	p1field3, ok := payload["field3"].(map[string]interface{})
	if !ok {
		t.Error("Fail to cast payload field3 to expected type")
	}
	p2field3, ok := payload2["field3"].(map[string]interface{})
	if !ok {
		t.Error("Fail to cast payload2 field3 to expected type")
	}
	p1field3["field4"] = 5.67
	if p2field3["field4"] != 1.02 {
		t.Errorf("unexpected field4 value in payload2 +%v\n", p2field3["field4"])
	}
}

func TestEventAck(t *testing.T) {
	ctx := context.Background()

	payload := map[string]interface{}{
		"field1": "abcd",
		"field2": 1234,
		"field3": map[string]interface{}{
			"field4": 1.02,
		},
	}

	done := make(chan bool)
	e1, err := event.New(ctx, payload, event.WithAck(
		func() {
			fmt.Println("success")
			done <- true
		},
		func(err error) {
			t.Errorf("Fail to receive all acknowledgements %s\n", err.Error())
			done <- true
		}))

	if err != nil {
		t.Errorf("Fail to create new event %s\n", err.Error())
	}

	e2, err := e1.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e3, err := e1.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e4, err := e2.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e5, err := e4.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e6, err := e4.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e7, err := e3.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e1.Ack()
	e2.Ack()
	e3.Ack()
	e4.Ack()
	e5.Ack()
	e6.Ack()
	e7.Ack()

	<-done

	//make sure we cannot do anything to events that are already acked
	var ackedErr *ack.AlreadyAckedError
	err = e1.SetPayload("blah")
	if err == nil || !errors.As(err, &ackedErr) {
		t.Errorf("Expect AlreadyAckedError but get +%v\n", err)
	}
	err = e5.SetContext(ctx)
	if err == nil || !errors.As(err, &ackedErr) {
		t.Errorf("Expect AlreadyAckedError but get +%v\n", err)
	}
	_, err = e7.Clone(ctx)
	if err == nil || !errors.As(err, &ackedErr) {
		t.Errorf("Expect AlreadyAckedError but get +%v\n", err)
	}
}

func TestEventNack(t *testing.T) {
	ctx := context.Background()

	payload := map[string]interface{}{
		"field1": "abcd",
		"field2": 1234,
		"field3": map[string]interface{}{
			"field4": 1.02,
		},
	}

	done := make(chan bool)
	e1, err := event.New(ctx, payload, event.WithAck(
		func() {
			t.Errorf("Expect error function to be called")
			done <- true
		},
		func(err error) {
			var nackErr *ack.NackError
			if !errors.As(err, &nackErr) {
				t.Errorf("Expect nackError but get +%v\n", err)
			}
			done <- true
		}))

	e2, err := e1.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e3, err := e1.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	e2.Nack(errors.New("error"))
	e3.Ack()

	<-done
}

func TestEventTimeout(t *testing.T) {
	ctx := context.Background()

	payload := map[string]interface{}{
		"field1": "abcd",
		"field2": 1234,
		"field3": map[string]interface{}{
			"field4": 1.02,
		},
	}

	ctx, _ = context.WithTimeout(ctx, time.Millisecond*500)

	done := make(chan bool)
	e1, err := event.New(ctx, payload, event.WithAck(
		func() {
			t.Errorf("Expect error function to be called")
			done <- true
		},
		func(err error) {
			var toErr *ack.TimeoutError
			if !errors.As(err, &toErr) {
				t.Errorf("Expect toErr but get +%v\n", err)
			}
			done <- true
		}))

	_, err = e1.Clone(ctx)
	if err != nil {
		t.Errorf("Fail to clone event %s\n", err.Error())
	}

	<-done
}
