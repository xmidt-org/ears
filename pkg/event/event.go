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
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/xmidt-org/ears/internal/pkg/ack"
	"github.com/xmidt-org/ears/internal/pkg/rtsemconv"
	"github.com/xmidt-org/ears/pkg/logs"
	"github.com/xmidt-org/ears/pkg/tenant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	//"github.com/mohae/deepcopy"
	//"github.com/gohobby/deepcopy"
	"github.com/boriwo/deepcopy"
)

type event struct {
	metadata           map[string]interface{}
	payload            interface{}
	ctx                context.Context
	ack                ack.SubTree
	tid                tenant.Id
	eid                string
	spanName           string
	span               trace.Span //Only valid in the root event
	tracePayloadOnNack bool
	tracePayload       interface{}
	userTraceId        string
	created            time.Time
	deepcopied         bool
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

// Create a new event given a context, a payload, and other event options
func New(ctx context.Context, payload interface{}, options ...EventOption) (Event, error) {
	e := &event{
		payload: payload,
		ctx:     ctx,
		ack:     nil,
		tid:     tenant.Id{OrgId: "", AppId: ""},
		created: time.Now(),
		eid:     uuid.New().String(),
	}
	for _, option := range options {
		err := option(e)
		if err != nil {
			return nil, err
		}
	}
	// enable otel tracing
	if e.spanName != "" {
		tracer := otel.Tracer(rtsemconv.EARSTracerName)
		var span trace.Span
		ctx, span = tracer.Start(ctx, e.spanName)
		span.SetAttributes(rtsemconv.EARSEventTrace)
		span.SetAttributes(rtsemconv.EARSOrgId.String(e.tid.OrgId), rtsemconv.EARSAppId.String(e.tid.AppId))
		if e.userTraceId == "" {
			e.userTraceId = span.SpanContext().TraceID().String()
		}
		span.SetAttributes(rtsemconv.EARSTraceId.String(e.userTraceId))
		e.span = span
	}
	// last resort, make up trace id here
	if e.userTraceId == "" {
		e.userTraceId = strings.Replace(uuid.New().String(), "-", "", -1)
	}
	// setting up logger for the event
	parentLogger, ok := logger.Load().(*zerolog.Logger)
	if ok {
		ctx = logs.SubLoggerCtx(ctx, parentLogger)
		logs.StrToLogCtx(ctx, rtsemconv.EarsLogTraceIdKey, e.userTraceId)
		logs.StrToLogCtx(ctx, rtsemconv.EarsLogTenantIdKey, e.tid.ToString())
		if e.span != nil {
			otelTraceId := e.span.SpanContext().TraceID().String()
			logs.StrToLogCtx(ctx, rtsemconv.EarsOtelTraceIdKey, otelTraceId)
		}
	}
	e.SetContext(ctx)
	return e, nil
}

// event acknowledge option with two completion functions,
// handledFn and errFn. An event with acknowledgement option will be notified through
// the handledFn when an event is handled, or through the errFn when there is an
// error handling it.
// An event is considered handled when it and all its child events (derived from the
// Clone function) have called the Ack function.
// An event is considered to have an error if it or any of its child events (derived from
// the Clone function) has called the Nack function.
// An event can also error out if it does not receive all the acknowledgements before
// the context timeout/cancellation.
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
				// log original payload here if desired
				if e.tracePayloadOnNack && e.tracePayload != nil {
					buf, err2 := json.Marshal(e.tracePayload)
					if err2 == nil {
						e.span.SetAttributes(attribute.String("payload", string(buf)))
					}
				}
				e.span.SetStatus(codes.Error, "event processing error")
				e.span.End()
			}
		})
		return nil
	}
}

func WithTracePayloadOnNack(tracePayloadOnNack bool) EventOption {
	return func(e *event) error {
		e.tracePayloadOnNack = tracePayloadOnNack
		if tracePayloadOnNack {
			e.tracePayload = deepcopy.DeepCopy(e.payload)
		}
		return nil
	}
}

func WithId(eid string) EventOption {
	return func(e *event) error {
		e.eid = eid
		return nil
	}
}

func WithUserTraceId(traceId string) EventOption {
	return func(e *event) error {
		e.userTraceId = traceId
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

// WithOtelTracing enables opentelemtry tracing for the event
func WithOtelTracing(spanName string) EventOption {
	return func(e *event) error {
		e.spanName = spanName
		return nil
	}
}

func (e *event) Created() time.Time {
	return e.created
}

func (e *event) Id() string {
	return e.eid
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

func (e *event) evalArrayPath(path string, curr interface{}) (interface{}, interface{}, string, int) {
	inArr := false
	last := 0
	segments := make([]string, 0)
	for pos, c := range path {
		if c == '[' {
			inArr = true
		} else if c == ']' {
			inArr = false
		}
		if c == '.' && !inArr && pos > last {
			segments = append(segments, path[(last+1):pos])
			last = pos
		} else if pos == len(path)-1 && pos-last > 0 {
			segments = append(segments, path[last+1:])
		}
	}
	var parent interface{}
	key := ""
	idx := -1
	for i := 0; i < len(segments); i++ {
		if curr == nil {
			return nil, nil, "", -1
		}
		elem := strings.TrimSpace(segments[i])
		if elem == "" {
			continue
		}
		curr, parent, key, idx = e.getChildElement(curr, elem)
	}
	return curr, parent, key, idx
}

func (e *event) getChildElement(curr interface{}, segment string) (interface{}, interface{}, string, int) {
	if curr == nil {
		return nil, nil, "", -1
	}
	if !strings.HasSuffix(segment, "]") {
		// map element
		switch curr := curr.(type) {
		case map[string]interface{}:
			return curr[segment], curr, segment, -1
		}
	} else {
		switch curr.(type) {
		case map[string]interface{}:
			// array element
			key := segment[:strings.Index(segment, "[")]
			switch curr.(map[string]interface{})[key].(type) {
			case []interface{}:
				idxStr := segment[strings.Index(segment, "[")+1 : len(segment)-1]
				if strings.Contains(idxStr, "=") {
					// array element by key selector
					kv := strings.Split(idxStr, "=")
					if len(kv) != 2 {
						return nil, nil, "", -1
					}
					k := strings.TrimSpace(kv[0])
					v := strings.TrimSpace(kv[1])
					for idx, ae := range curr.(map[string]interface{})[key].([]interface{}) {
						switch ae := ae.(type) {
						case map[string]interface{}:
							if ae[k] == v {
								return ae, curr, key, idx
							}
						}
					}
				} else {
					// array element by index
					idx, err := strconv.Atoi(idxStr)
					if err != nil {
						return nil, nil, "", -1
					}
					if idx >= len(curr.(map[string]interface{})[key].([]interface{})) {
						//ctx.Log().Error("error_type", "parser", "cause", "array_length_error", "path", path)
						return nil, nil, "", -1
					}
					return curr.(map[string]interface{})[key].([]interface{})[idx], curr, key, idx
				}
			}
		case []interface{}:
			idxStr := segment[strings.Index(segment, "[")+1 : len(segment)-1]
			if strings.Contains(idxStr, "=") {
				// array element by key selector
				kv := strings.Split(idxStr, "=")
				if len(kv) != 2 {
					return nil, nil, "", -1
				}
				k := strings.TrimSpace(kv[0])
				v := strings.TrimSpace(kv[1])
				for idx, ae := range curr.([]interface{}) {
					switch ae := ae.(type) {
					case map[string]interface{}:
						if ae[k] == v {
							return ae, curr, "", idx
						}
					}
				}
			} else {
				// array element by index
				idx, err := strconv.Atoi(idxStr)
				if err != nil {
					return nil, nil, "", -1
				}
				if idx >= len(curr.([]interface{})) {
					//ctx.Log().Error("error_type", "parser", "cause", "array_length_error", "path", path)
					return nil, nil, "", -1
				}
				return curr.([]interface{})[idx], curr, "", idx
			}
		}
	}
	return nil, nil, "", -1
}

func (e *event) Evaluate(expression interface{}) (interface{}, interface{}, string) {
	switch expr := expression.(type) {
	case string:
		for {
			si := strings.Index(expr, "{")
			ei := strings.Index(expr, "}")
			if si < 0 || ei < 0 {
				break
			}
			path := expr[si+1 : ei]
			v, pr, ky := e.GetPathValue(path)
			v = deepcopy.DeepCopy(v)
			if si == 0 && ei == len(expr)-1 {
				return v, pr, ky
			}
			switch vt := v.(type) {
			case string:
				expr = expr[0:si] + vt + expr[ei+1:]
			default:
				sv, _ := json.Marshal(vt)
				expr = expr[0:si] + string(sv) + expr[ei+1:]
			}
		}
		return expr, nil, ""
	default:
		return expression, nil, ""
	}
	return expression, nil, ""
}

func (e *event) splitPath(path string) []string {
	path = strings.Replace(path, `\.`, `\\`, -1)
	segments := strings.Split(path, ".")
	for i, _ := range segments {
		segments[i] = strings.Replace(segments[i], `\\`, `.`, -1)
	}
	return segments
}

func (e *event) GetPathValue(path string) (interface{}, interface{}, string) {
	if path == TRACE+".id" {
		if e.userTraceId != "" {
			return e.userTraceId, nil, ""
		}
		traceId := trace.SpanFromContext(e.ctx).SpanContext().TraceID().String()
		return traceId, nil, ""
	}
	if path == TENANT+".appId" {
		return e.Tenant().AppId, nil, ""
	}
	if path == TENANT+".orgId" {
		return e.Tenant().OrgId, nil, ""
	}
	if path == TIMESTAMP {
		return strconv.Itoa(int(time.Now().UnixNano() / 1e6)), nil, ""
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
	if strings.Contains(path, "[") && strings.Contains(path, "]") {
		v, p, k, _ := e.evalArrayPath(path[strings.Index(path, "."):], obj)
		return v, p, k
	}
	var parent interface{}
	var key string
	var ok bool
	segments := e.splitPath(path)
	if len(segments) < 2 {
		return nil, nil, ""
	}
	for i := 1; i < len(segments); i++ {
		s := segments[i]
		parent = obj
		key = s
		switch tobj := obj.(type) {
		case map[string]interface{}:
			obj, ok = tobj[s]
			if !ok {
				return nil, parent, key
			}
		default:
			return nil, parent, key
		}
	}
	return obj, parent, key
}

func (e *event) SetPathValue(path string, val interface{}, createPath bool) (interface{}, string, error) {
	if e.ack != nil && e.ack.IsAcked() {
		return nil, "", &ack.AlreadyAckedError{}
	}
	obj := e.Payload()
	if strings.HasPrefix(path, METADATA+".") || path == METADATA {
		metaObj := e.Metadata()
		if metaObj == nil {
			if createPath {
				e.SetMetadata(make(map[string]interface{}))
				obj = e.Metadata()
			} else {
				return nil, "", errors.New("path " + path + " does not exist")
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
	if path == PAYLOAD || path == PAYLOAD+"." {
		err := e.SetPayload(val)
		return nil, "", err
	} else if path == METADATA || path == METADATA+"." {
		valMap, ok := val.(map[string]interface{})
		if ok {
			e.SetMetadata(valMap)
		}
		return nil, "", errors.New("bad metadata")
	}
	if !strings.HasPrefix(path, PAYLOAD+".") && !strings.HasPrefix(path, METADATA+".") {
		return nil, "", errors.New("bad path " + path)
	}
	var parent interface{}
	var key string
	var ok bool
	segments := e.splitPath(path)
	for i := 1; i < len(segments); i++ {
		s := segments[i]
		if s == "" {
			continue
		}
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
				return nil, "", errors.New("invalid path " + path)
			}
		}
	}
	switch obj := obj.(type) {
	case map[string]interface{}:
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			idx, _ := strconv.Atoi(key[strings.Index(key, "[")+1 : strings.Index(key, "]")])
			key = key[:strings.Index(key, "[")]
			if idx >= 0 {
				if obj[key] == nil && createPath {
					obj[key] = make([]interface{}, idx+1)
				}
				if idx < len(obj[key].([]interface{})) {
					obj[key].([]interface{})[idx] = val
				} else {
					// increase array size
					newArr := make([]interface{}, idx+1)
					for i := 0; i < idx; i++ {
						newArr[i] = obj[key].([]interface{})[i]
					}
					newArr[idx] = val
					obj[key] = newArr
				}
			}
		} else {
			obj[key] = val
		}
	}
	return parent, key, nil
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

func (e *event) UserTraceId() string {
	return e.userTraceId
}

func (e *event) Clone(ctx context.Context) (Event, error) {
	// clone shallow
	var subTree ack.SubTree
	if e.ack != nil {
		var err error
		subTree, err = e.ack.NewSubTree()
		if err != nil {
			return nil, err
		}
	}
	return &event{
		payload:     e.Payload(),
		metadata:    e.Metadata(),
		ctx:         ctx,
		ack:         subTree,
		eid:         e.eid,
		tid:         e.tid,
		created:     e.created,
		userTraceId: e.userTraceId,
	}, nil
}

func (e *event) DeepCopy() error {
	// only deepcopy once per clone
	if e.deepcopied {
		return nil
	}
	e.deepcopied = true
	// gohobby deepcopy without reflection
	// benchmark for 1.7 kb event: 6,000 ns / op
	e.payload = deepcopy.DeepCopy(e.payload)
	e.metadata = deepcopy.DeepCopy(e.metadata).(map[string]interface{})
	return nil
}

/*func (e *event) DeepCopy() error {
	if e.deepcopied {
		return nil
	}
	e.deepcopied = true
	// mohae deepcopy using reflection
	// very fast according to the following benchmark:
	// https://xuri.me/2018/06/17/deep-copy-object-with-reflecting-or-gob-in-go.html
	// unclear if all types are supported:
	// https://github.com/mohae/deepcopy/blob/master/deepcopy.go#L45-L46
	// benchmark for 1.7 kb event: 29,000 ns / op
	e.payload = deepcopy.Copy(e.payload)
	e.metadata = deepcopy.Copy(e.metadata).(map[string]interface{})
	return nil
}*/

/*func (e *event) DeepCopy() error {
	if e.deepcopied {
		return nil
	}
	e.deepcopied = true
	// trivial json serialization / deserialization implementation
	// benchmark for 1.7 kb event: 35,000 ns / op
	buf, err := json.Marshal(e.Payload())
	if err != nil {
		return err
	}
	var payload interface{}
	err = json.Unmarshal(buf, &payload)
	if err != nil {
		return err
	}
	buf, err = json.Marshal(e.Metadata())
	if err != nil {
		return err
	}
	var metadata map[string]interface{}
	err = json.Unmarshal(buf, &metadata)
	if err != nil {
		return err
	}
	e.payload = payload
	e.metadata = metadata
	return nil
}*/
