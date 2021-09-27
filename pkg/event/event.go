// Licensed to Comcast Cable Communications Management, LLC under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Comcast Cable Communications Management, LLC licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package event

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/ack"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/logs"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mohae/deepcopy"
)

type event struct {
	metadata map[string]interface{}
	payload  interface{}
	ctx      context.Context
	ack      ack.SubTree
	tid      tenant.Id
	spanName string
	span     trace.Span //Only valid in the root event
	created  time.Time
}

type EventOption func(*event) error

var logger atomic.Value

func SetEventLogger(l *zerolog.Logger) {
	logger.Store(l)
}

func GetEventLogger() *zerolog.Logger {
	parentLogger, ok := logger.Load().(*zerolog.Logger)
	if ok {
		return parentLogger
	}
	return nil
}

//Create a new event given a context, a payload, and other event options
func New(ctx context.Context, payload interface{}, options ...EventOption) (Event, error) {
	e := &event{
		payload: payload,
		ctx:     ctx,
		ack:     nil,
		tid:     tenant.Id{OrgId: "", AppId: ""},
		created: time.Now(),
	}
	for _, option := range options {
		err := option(e)
		if err != nil {
			return nil, err
		}
	}
	traceId := uuid.New().String()

	// enable otel tracing
	if e.spanName != "" {
		tracer := otel.Tracer(rtsemconv.EARSTracerName)
		var span trace.Span
		ctx, span = tracer.Start(ctx, e.spanName)
		span.SetAttributes(rtsemconv.EARSEventTrace)
		span.SetAttributes(rtsemconv.EARSOrgId.String(e.tid.OrgId), rtsemconv.EARSAppId.String(e.tid.AppId))
		traceId = span.SpanContext().TraceID().String()
		span.SetAttributes(rtsemconv.EARSTraceId.String(traceId))
		e.span = span
	}

	// setting up logger for the event
	parentLogger, ok := logger.Load().(*zerolog.Logger)
	if ok {
		ctx = logs.SubLoggerCtx(ctx, parentLogger)
		logs.StrToLogCtx(ctx, rtsemconv.EarsLogTraceIdKey, traceId)
		logs.StrToLogCtx(ctx, rtsemconv.EarsLogTenantIdKey, e.tid.ToString())
	}
	e.SetContext(ctx)
	return e, nil
}

//event acknowledge option with two completion functions,
//handledFn and errFn. An event with acknowledgement option will be notified through
//the handledFn when an event is handled, or through the errFn when there is an
//error handling it.
//An event is considered handled when it and all its child events (derived from the
//Clone function) have called the Ack function.
//An event is considered to have an error if it or any of its child events (derived from
//the Clone function) has called the Nack function.
//An event can also error out if it does not receive all the acknowledgements before
//the context timeout/cancellation.
func WithAck(handledFn func(Event), errFn func(Event, error)) EventOption {
	return func(e *event) error {
		if handledFn == nil || errFn == nil {
			return &NoAckHandlersError{}
		}
		e.ack = ack.NewAckTree(e.ctx, func() {
			handledFn(e)
			if e.span != nil {
				e.span.AddEvent("ack")
				e.span.End()
			}
		}, func(err error) {
			errFn(e, err)
			if e.span != nil {
				e.span.AddEvent("nack")
				e.span.RecordError(err)
				e.span.SetStatus(codes.Error, "event processing error")
				e.span.End()
			}
		})
		return nil
	}
}

func WithMetadata(metadata map[string]interface{}) EventOption {
	return func(e *event) error {
		e.SetMetadata(metadata)
		return nil
	}
}

func WithMetadataKeyValue(key string, val interface{}) EventOption {
	return func(e *event) error {
		if e.metadata == nil {
			e.metadata = make(map[string]interface{})
		}
		e.metadata[key] = val
		return nil
	}
}

func WithTenant(tid tenant.Id) EventOption {
	return func(e *event) error {
		e.tid = tid
		return nil
	}
}

//WithOtelTracing enables opentelemtry tracing for the event
func WithOtelTracing(spanName string) EventOption {
	return func(e *event) error {
		e.spanName = spanName
		return nil
	}
}

func (e *event) Created() time.Time {
	return e.created
}

func (e *event) Payload() interface{} {
	return e.payload
}

func (e *event) SetPayload(payload interface{}) error {
	if e.ack != nil && e.ack.IsAcked() {
		return &ack.AlreadyAckedError{}
	}
	e.payload = payload
	return nil
}

func (e *event) Metadata() map[string]interface{} {
	return e.metadata
}

func (e *event) SetMetadata(metadata map[string]interface{}) error {
	if e.ack != nil && e.ack.IsAcked() {
		return &ack.AlreadyAckedError{}
	}
	e.metadata = metadata
	return nil
}

func (e *event) Tenant() tenant.Id {
	return e.tid
}

func (e *event) GetPathValue(path string) (interface{}, interface{}, string) {
	if path == TRACE+".id" {
		traceId := trace.SpanFromContext(e.ctx).SpanContext().TraceID().String()
		return traceId, nil, ""
	}
	if path == TENANT+".appId" {
		return e.Tenant().AppId, nil, ""
	}
	if path == TENANT+".orgId" {
		return e.Tenant().OrgId, nil, ""
	}
	obj := e.Payload()
	if strings.HasPrefix(path, METADATA+".") || path == METADATA {
		obj = e.Metadata()
	}
	if obj == nil {
		return nil, nil, ""
	}
	if path == "" || path == "." || path == PAYLOAD || path == METADATA {
		return obj, nil, ""
	}
	if strings.HasPrefix(path, ".") {
		path = PAYLOAD + path
	}
	if !strings.HasPrefix(path, PAYLOAD+".") && !strings.HasPrefix(path, METADATA+".") {
		return nil, nil, ""
	}
	var parent interface{}
	var key string
	var ok bool
	segments := strings.Split(path, ".")
	if len(segments) < 2 {
		return nil, nil, ""
	}
	for i := 1; i < len(segments); i++ {
		s := segments[i]
		parent = obj
		key = s
		obj, ok = obj.(map[string]interface{})[s]
		if !ok {
			return nil, parent, key
		}
	}
	return obj, parent, key
}

func (e *event) SetPathValue(path string, val interface{}, createPath bool) (interface{}, string) {
	obj := e.Payload()
	if strings.HasPrefix(path, METADATA+".") || path == METADATA {
		metaObj := e.Metadata()
		if metaObj == nil {
			if createPath {
				e.SetMetadata(make(map[string]interface{}))
				obj = e.Metadata()
			} else {
				return nil, ""
			}
		} else {
			obj = metaObj
		}
	}
	if strings.HasPrefix(path, ".") {
		path = PAYLOAD + path
	}
	if path == "" {
		path = PAYLOAD
	}
	if path == PAYLOAD {
		e.SetPayload(val)
		return nil, ""
	} else if path == METADATA {
		valMap, ok := val.(map[string]interface{})
		if ok {
			e.SetMetadata(valMap)
		}
		return nil, ""
	}
	if !strings.HasPrefix(path, PAYLOAD+".") && !strings.HasPrefix(path, METADATA+".") {
		return nil, ""
	}
	var parent interface{}
	var key string
	var ok bool
	segments := strings.Split(path, ".")
	for i := 1; i < len(segments); i++ {
		s := segments[i]
		parent = obj
		key = s
		if i == len(segments)-1 {
			break
		}
		obj, ok = obj.(map[string]interface{})[s]
		if !ok {
			if createPath && i < len(segments)-1 {
				parent.(map[string]interface{})[key] = make(map[string]interface{})
				obj = parent.(map[string]interface{})[key]
			} else {
				return nil, ""
			}
		}
	}
	obj.(map[string]interface{})[key] = val
	return parent, key
}

func (e *event) Context() context.Context {
	return e.ctx
}

func (e *event) SetContext(ctx context.Context) error {
	if e.ack != nil && e.ack.IsAcked() {
		return &ack.AlreadyAckedError{}
	}
	e.ctx = ctx
	return nil
}

func (e *event) Ack() {
	if e.ack != nil {
		e.ack.Ack()
	}
}

func (e *event) Nack(err error) {
	if e.ack != nil {
		e.ack.Nack(err)
	}
}

func (e *event) Clone(ctx context.Context) (Event, error) {
	var subTree ack.SubTree
	if e.ack != nil {
		var err error
		subTree, err = e.ack.NewSubTree()
		if err != nil {
			return nil, err
		}
	}
	//Very fast according to the following benchmark:
	//https://xuri.me/2018/06/17/deep-copy-object-with-reflecting-or-gob-in-go.html
	//Unclear if all types are supported:
	//https://github.com/mohae/deepcopy/blob/master/deepcopy.go#L45-L46
	newPayloadCopy := deepcopy.Copy(e.payload)
	newMetadtaCopy := deepcopy.Copy(e.metadata).(map[string]interface{})
	return &event{
		payload:  newPayloadCopy,
		metadata: newMetadtaCopy,
		ctx:      ctx,
		ack:      subTree,
		tid:      e.tid,
		created:  e.created,
	}, nil
}
