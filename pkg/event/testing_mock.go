// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package event

import (
	"context"
	"github.com/xmidt-org/ears/pkg/tenant"
	"sync"
	"time"
)

// Ensure, that EventMock does implement Event.
// If this is not the case, regenerate this file with moq.
var _ Event = &EventMock{}

// EventMock is a mock implementation of Event.
//
//	func TestSomethingThatUsesEvent(t *testing.T) {
//
//		// make and configure a mocked Event
//		mockedEvent := &EventMock{
//			AckFunc: func()  {
//				panic("mock out the Ack method")
//			},
//			CloneFunc: func(ctx context.Context) (Event, error) {
//				panic("mock out the Clone method")
//			},
//			ContextFunc: func() context.Context {
//				panic("mock out the Context method")
//			},
//			CreatedFunc: func() time.Time {
//				panic("mock out the Created method")
//			},
//			DeepCopyFunc: func() error {
//				panic("mock out the DeepCopy method")
//			},
//			EvaluateFunc: func(expression interface{}) (interface{}, interface{}, string) {
//				panic("mock out the Evaluate method")
//			},
//			GetPathValueFunc: func(path string) (interface{}, interface{}, string) {
//				panic("mock out the GetPathValue method")
//			},
//			IdFunc: func() string {
//				panic("mock out the Id method")
//			},
//			MetadataFunc: func() map[string]interface{} {
//				panic("mock out the Metadata method")
//			},
//			NackFunc: func(err error)  {
//				panic("mock out the Nack method")
//			},
//			PayloadFunc: func() interface{} {
//				panic("mock out the Payload method")
//			},
//			ResponseFunc: func() string {
//				panic("mock out the Response method")
//			},
//			SetContextFunc: func(ctx context.Context) error {
//				panic("mock out the SetContext method")
//			},
//			SetMetadataFunc: func(metadata map[string]interface{}) error {
//				panic("mock out the SetMetadata method")
//			},
//			SetPathValueFunc: func(path string, val interface{}, createPath bool) (interface{}, string, error) {
//				panic("mock out the SetPathValue method")
//			},
//			SetPayloadFunc: func(payload interface{}) error {
//				panic("mock out the SetPayload method")
//			},
//			SetResponseFunc: func(response string)  {
//				panic("mock out the SetResponse method")
//			},
//			TenantFunc: func() tenant.Id {
//				panic("mock out the Tenant method")
//			},
//			UserTraceIdFunc: func() string {
//				panic("mock out the UserTraceId method")
//			},
//		}
//
//		// use mockedEvent in code that requires Event
//		// and then make assertions.
//
//	}
type EventMock struct {
	// AckFunc mocks the Ack method.
	AckFunc func()

	// CloneFunc mocks the Clone method.
	CloneFunc func(ctx context.Context) (Event, error)

	// ContextFunc mocks the Context method.
	ContextFunc func() context.Context

	// CreatedFunc mocks the Created method.
	CreatedFunc func() time.Time

	// DeepCopyFunc mocks the DeepCopy method.
	DeepCopyFunc func() error

	// EvaluateFunc mocks the Evaluate method.
	EvaluateFunc func(expression interface{}) (interface{}, interface{}, string)

	// GetPathValueFunc mocks the GetPathValue method.
	GetPathValueFunc func(path string) (interface{}, interface{}, string)

	// IdFunc mocks the Id method.
	IdFunc func() string

	// MetadataFunc mocks the Metadata method.
	MetadataFunc func() map[string]interface{}

	// NackFunc mocks the Nack method.
	NackFunc func(err error)

	// PayloadFunc mocks the Payload method.
	PayloadFunc func() interface{}

	// ResponseFunc mocks the Response method.
	ResponseFunc func() string

	// SetContextFunc mocks the SetContext method.
	SetContextFunc func(ctx context.Context) error

	// SetMetadataFunc mocks the SetMetadata method.
	SetMetadataFunc func(metadata map[string]interface{}) error

	// SetPathValueFunc mocks the SetPathValue method.
	SetPathValueFunc func(path string, val interface{}, createPath bool) (interface{}, string, error)

	// SetPayloadFunc mocks the SetPayload method.
	SetPayloadFunc func(payload interface{}) error

	// SetResponseFunc mocks the SetResponse method.
	SetResponseFunc func(response string)

	// TenantFunc mocks the Tenant method.
	TenantFunc func() tenant.Id

	// UserTraceIdFunc mocks the UserTraceId method.
	UserTraceIdFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Ack holds details about calls to the Ack method.
		Ack []struct {
		}
		// Clone holds details about calls to the Clone method.
		Clone []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Context holds details about calls to the Context method.
		Context []struct {
		}
		// Created holds details about calls to the Created method.
		Created []struct {
		}
		// DeepCopy holds details about calls to the DeepCopy method.
		DeepCopy []struct {
		}
		// Evaluate holds details about calls to the Evaluate method.
		Evaluate []struct {
			// Expression is the expression argument value.
			Expression interface{}
		}
		// GetPathValue holds details about calls to the GetPathValue method.
		GetPathValue []struct {
			// Path is the path argument value.
			Path string
		}
		// Id holds details about calls to the Id method.
		Id []struct {
		}
		// Metadata holds details about calls to the Metadata method.
		Metadata []struct {
		}
		// Nack holds details about calls to the Nack method.
		Nack []struct {
			// Err is the err argument value.
			Err error
		}
		// Payload holds details about calls to the Payload method.
		Payload []struct {
		}
		// Response holds details about calls to the Response method.
		Response []struct {
		}
		// SetContext holds details about calls to the SetContext method.
		SetContext []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// SetMetadata holds details about calls to the SetMetadata method.
		SetMetadata []struct {
			// Metadata is the metadata argument value.
			Metadata map[string]interface{}
		}
		// SetPathValue holds details about calls to the SetPathValue method.
		SetPathValue []struct {
			// Path is the path argument value.
			Path string
			// Val is the val argument value.
			Val interface{}
			// CreatePath is the createPath argument value.
			CreatePath bool
		}
		// SetPayload holds details about calls to the SetPayload method.
		SetPayload []struct {
			// Payload is the payload argument value.
			Payload interface{}
		}
		// SetResponse holds details about calls to the SetResponse method.
		SetResponse []struct {
			// Response is the response argument value.
			Response string
		}
		// Tenant holds details about calls to the Tenant method.
		Tenant []struct {
		}
		// UserTraceId holds details about calls to the UserTraceId method.
		UserTraceId []struct {
		}
	}
	lockAck          sync.RWMutex
	lockClone        sync.RWMutex
	lockContext      sync.RWMutex
	lockCreated      sync.RWMutex
	lockDeepCopy     sync.RWMutex
	lockEvaluate     sync.RWMutex
	lockGetPathValue sync.RWMutex
	lockId           sync.RWMutex
	lockMetadata     sync.RWMutex
	lockNack         sync.RWMutex
	lockPayload      sync.RWMutex
	lockResponse     sync.RWMutex
	lockSetContext   sync.RWMutex
	lockSetMetadata  sync.RWMutex
	lockSetPathValue sync.RWMutex
	lockSetPayload   sync.RWMutex
	lockSetResponse  sync.RWMutex
	lockTenant       sync.RWMutex
	lockUserTraceId  sync.RWMutex
}

// Ack calls AckFunc.
func (mock *EventMock) Ack() {
	if mock.AckFunc == nil {
		panic("EventMock.AckFunc: method is nil but Event.Ack was just called")
	}
	callInfo := struct {
	}{}
	mock.lockAck.Lock()
	mock.calls.Ack = append(mock.calls.Ack, callInfo)
	mock.lockAck.Unlock()
	mock.AckFunc()
}

// AckCalls gets all the calls that were made to Ack.
// Check the length with:
//
//	len(mockedEvent.AckCalls())
func (mock *EventMock) AckCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockAck.RLock()
	calls = mock.calls.Ack
	mock.lockAck.RUnlock()
	return calls
}

// Clone calls CloneFunc.
func (mock *EventMock) Clone(ctx context.Context) (Event, error) {
	if mock.CloneFunc == nil {
		panic("EventMock.CloneFunc: method is nil but Event.Clone was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockClone.Lock()
	mock.calls.Clone = append(mock.calls.Clone, callInfo)
	mock.lockClone.Unlock()
	return mock.CloneFunc(ctx)
}

// CloneCalls gets all the calls that were made to Clone.
// Check the length with:
//
//	len(mockedEvent.CloneCalls())
func (mock *EventMock) CloneCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockClone.RLock()
	calls = mock.calls.Clone
	mock.lockClone.RUnlock()
	return calls
}

// Context calls ContextFunc.
func (mock *EventMock) Context() context.Context {
	if mock.ContextFunc == nil {
		panic("EventMock.ContextFunc: method is nil but Event.Context was just called")
	}
	callInfo := struct {
	}{}
	mock.lockContext.Lock()
	mock.calls.Context = append(mock.calls.Context, callInfo)
	mock.lockContext.Unlock()
	return mock.ContextFunc()
}

// ContextCalls gets all the calls that were made to Context.
// Check the length with:
//
//	len(mockedEvent.ContextCalls())
func (mock *EventMock) ContextCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockContext.RLock()
	calls = mock.calls.Context
	mock.lockContext.RUnlock()
	return calls
}

// Created calls CreatedFunc.
func (mock *EventMock) Created() time.Time {
	if mock.CreatedFunc == nil {
		panic("EventMock.CreatedFunc: method is nil but Event.Created was just called")
	}
	callInfo := struct {
	}{}
	mock.lockCreated.Lock()
	mock.calls.Created = append(mock.calls.Created, callInfo)
	mock.lockCreated.Unlock()
	return mock.CreatedFunc()
}

// CreatedCalls gets all the calls that were made to Created.
// Check the length with:
//
//	len(mockedEvent.CreatedCalls())
func (mock *EventMock) CreatedCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockCreated.RLock()
	calls = mock.calls.Created
	mock.lockCreated.RUnlock()
	return calls
}

// DeepCopy calls DeepCopyFunc.
func (mock *EventMock) DeepCopy() error {
	if mock.DeepCopyFunc == nil {
		panic("EventMock.DeepCopyFunc: method is nil but Event.DeepCopy was just called")
	}
	callInfo := struct {
	}{}
	mock.lockDeepCopy.Lock()
	mock.calls.DeepCopy = append(mock.calls.DeepCopy, callInfo)
	mock.lockDeepCopy.Unlock()
	return mock.DeepCopyFunc()
}

// DeepCopyCalls gets all the calls that were made to DeepCopy.
// Check the length with:
//
//	len(mockedEvent.DeepCopyCalls())
func (mock *EventMock) DeepCopyCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockDeepCopy.RLock()
	calls = mock.calls.DeepCopy
	mock.lockDeepCopy.RUnlock()
	return calls
}

// Evaluate calls EvaluateFunc.
func (mock *EventMock) Evaluate(expression interface{}) (interface{}, interface{}, string) {
	if mock.EvaluateFunc == nil {
		panic("EventMock.EvaluateFunc: method is nil but Event.Evaluate was just called")
	}
	callInfo := struct {
		Expression interface{}
	}{
		Expression: expression,
	}
	mock.lockEvaluate.Lock()
	mock.calls.Evaluate = append(mock.calls.Evaluate, callInfo)
	mock.lockEvaluate.Unlock()
	return mock.EvaluateFunc(expression)
}

// EvaluateCalls gets all the calls that were made to Evaluate.
// Check the length with:
//
//	len(mockedEvent.EvaluateCalls())
func (mock *EventMock) EvaluateCalls() []struct {
	Expression interface{}
} {
	var calls []struct {
		Expression interface{}
	}
	mock.lockEvaluate.RLock()
	calls = mock.calls.Evaluate
	mock.lockEvaluate.RUnlock()
	return calls
}

// GetPathValue calls GetPathValueFunc.
func (mock *EventMock) GetPathValue(path string) (interface{}, interface{}, string) {
	if mock.GetPathValueFunc == nil {
		panic("EventMock.GetPathValueFunc: method is nil but Event.GetPathValue was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockGetPathValue.Lock()
	mock.calls.GetPathValue = append(mock.calls.GetPathValue, callInfo)
	mock.lockGetPathValue.Unlock()
	return mock.GetPathValueFunc(path)
}

// GetPathValueCalls gets all the calls that were made to GetPathValue.
// Check the length with:
//
//	len(mockedEvent.GetPathValueCalls())
func (mock *EventMock) GetPathValueCalls() []struct {
	Path string
} {
	var calls []struct {
		Path string
	}
	mock.lockGetPathValue.RLock()
	calls = mock.calls.GetPathValue
	mock.lockGetPathValue.RUnlock()
	return calls
}

// Id calls IdFunc.
func (mock *EventMock) Id() string {
	if mock.IdFunc == nil {
		panic("EventMock.IdFunc: method is nil but Event.Id was just called")
	}
	callInfo := struct {
	}{}
	mock.lockId.Lock()
	mock.calls.Id = append(mock.calls.Id, callInfo)
	mock.lockId.Unlock()
	return mock.IdFunc()
}

// IdCalls gets all the calls that were made to Id.
// Check the length with:
//
//	len(mockedEvent.IdCalls())
func (mock *EventMock) IdCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockId.RLock()
	calls = mock.calls.Id
	mock.lockId.RUnlock()
	return calls
}

// Metadata calls MetadataFunc.
func (mock *EventMock) Metadata() map[string]interface{} {
	if mock.MetadataFunc == nil {
		panic("EventMock.MetadataFunc: method is nil but Event.Metadata was just called")
	}
	callInfo := struct {
	}{}
	mock.lockMetadata.Lock()
	mock.calls.Metadata = append(mock.calls.Metadata, callInfo)
	mock.lockMetadata.Unlock()
	return mock.MetadataFunc()
}

// MetadataCalls gets all the calls that were made to Metadata.
// Check the length with:
//
//	len(mockedEvent.MetadataCalls())
func (mock *EventMock) MetadataCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockMetadata.RLock()
	calls = mock.calls.Metadata
	mock.lockMetadata.RUnlock()
	return calls
}

// Nack calls NackFunc.
func (mock *EventMock) Nack(err error) {
	if mock.NackFunc == nil {
		panic("EventMock.NackFunc: method is nil but Event.Nack was just called")
	}
	callInfo := struct {
		Err error
	}{
		Err: err,
	}
	mock.lockNack.Lock()
	mock.calls.Nack = append(mock.calls.Nack, callInfo)
	mock.lockNack.Unlock()
	mock.NackFunc(err)
}

// NackCalls gets all the calls that were made to Nack.
// Check the length with:
//
//	len(mockedEvent.NackCalls())
func (mock *EventMock) NackCalls() []struct {
	Err error
} {
	var calls []struct {
		Err error
	}
	mock.lockNack.RLock()
	calls = mock.calls.Nack
	mock.lockNack.RUnlock()
	return calls
}

// Payload calls PayloadFunc.
func (mock *EventMock) Payload() interface{} {
	if mock.PayloadFunc == nil {
		panic("EventMock.PayloadFunc: method is nil but Event.Payload was just called")
	}
	callInfo := struct {
	}{}
	mock.lockPayload.Lock()
	mock.calls.Payload = append(mock.calls.Payload, callInfo)
	mock.lockPayload.Unlock()
	return mock.PayloadFunc()
}

// PayloadCalls gets all the calls that were made to Payload.
// Check the length with:
//
//	len(mockedEvent.PayloadCalls())
func (mock *EventMock) PayloadCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPayload.RLock()
	calls = mock.calls.Payload
	mock.lockPayload.RUnlock()
	return calls
}

// Response calls ResponseFunc.
func (mock *EventMock) Response() string {
	if mock.ResponseFunc == nil {
		panic("EventMock.ResponseFunc: method is nil but Event.Response was just called")
	}
	callInfo := struct {
	}{}
	mock.lockResponse.Lock()
	mock.calls.Response = append(mock.calls.Response, callInfo)
	mock.lockResponse.Unlock()
	return mock.ResponseFunc()
}

// ResponseCalls gets all the calls that were made to Response.
// Check the length with:
//
//	len(mockedEvent.ResponseCalls())
func (mock *EventMock) ResponseCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockResponse.RLock()
	calls = mock.calls.Response
	mock.lockResponse.RUnlock()
	return calls
}

// SetContext calls SetContextFunc.
func (mock *EventMock) SetContext(ctx context.Context) error {
	if mock.SetContextFunc == nil {
		panic("EventMock.SetContextFunc: method is nil but Event.SetContext was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockSetContext.Lock()
	mock.calls.SetContext = append(mock.calls.SetContext, callInfo)
	mock.lockSetContext.Unlock()
	return mock.SetContextFunc(ctx)
}

// SetContextCalls gets all the calls that were made to SetContext.
// Check the length with:
//
//	len(mockedEvent.SetContextCalls())
func (mock *EventMock) SetContextCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockSetContext.RLock()
	calls = mock.calls.SetContext
	mock.lockSetContext.RUnlock()
	return calls
}

// SetMetadata calls SetMetadataFunc.
func (mock *EventMock) SetMetadata(metadata map[string]interface{}) error {
	if mock.SetMetadataFunc == nil {
		panic("EventMock.SetMetadataFunc: method is nil but Event.SetMetadata was just called")
	}
	callInfo := struct {
		Metadata map[string]interface{}
	}{
		Metadata: metadata,
	}
	mock.lockSetMetadata.Lock()
	mock.calls.SetMetadata = append(mock.calls.SetMetadata, callInfo)
	mock.lockSetMetadata.Unlock()
	return mock.SetMetadataFunc(metadata)
}

// SetMetadataCalls gets all the calls that were made to SetMetadata.
// Check the length with:
//
//	len(mockedEvent.SetMetadataCalls())
func (mock *EventMock) SetMetadataCalls() []struct {
	Metadata map[string]interface{}
} {
	var calls []struct {
		Metadata map[string]interface{}
	}
	mock.lockSetMetadata.RLock()
	calls = mock.calls.SetMetadata
	mock.lockSetMetadata.RUnlock()
	return calls
}

// SetPathValue calls SetPathValueFunc.
func (mock *EventMock) SetPathValue(path string, val interface{}, createPath bool) (interface{}, string, error) {
	if mock.SetPathValueFunc == nil {
		panic("EventMock.SetPathValueFunc: method is nil but Event.SetPathValue was just called")
	}
	callInfo := struct {
		Path       string
		Val        interface{}
		CreatePath bool
	}{
		Path:       path,
		Val:        val,
		CreatePath: createPath,
	}
	mock.lockSetPathValue.Lock()
	mock.calls.SetPathValue = append(mock.calls.SetPathValue, callInfo)
	mock.lockSetPathValue.Unlock()
	return mock.SetPathValueFunc(path, val, createPath)
}

// SetPathValueCalls gets all the calls that were made to SetPathValue.
// Check the length with:
//
//	len(mockedEvent.SetPathValueCalls())
func (mock *EventMock) SetPathValueCalls() []struct {
	Path       string
	Val        interface{}
	CreatePath bool
} {
	var calls []struct {
		Path       string
		Val        interface{}
		CreatePath bool
	}
	mock.lockSetPathValue.RLock()
	calls = mock.calls.SetPathValue
	mock.lockSetPathValue.RUnlock()
	return calls
}

// SetPayload calls SetPayloadFunc.
func (mock *EventMock) SetPayload(payload interface{}) error {
	if mock.SetPayloadFunc == nil {
		panic("EventMock.SetPayloadFunc: method is nil but Event.SetPayload was just called")
	}
	callInfo := struct {
		Payload interface{}
	}{
		Payload: payload,
	}
	mock.lockSetPayload.Lock()
	mock.calls.SetPayload = append(mock.calls.SetPayload, callInfo)
	mock.lockSetPayload.Unlock()
	return mock.SetPayloadFunc(payload)
}

// SetPayloadCalls gets all the calls that were made to SetPayload.
// Check the length with:
//
//	len(mockedEvent.SetPayloadCalls())
func (mock *EventMock) SetPayloadCalls() []struct {
	Payload interface{}
} {
	var calls []struct {
		Payload interface{}
	}
	mock.lockSetPayload.RLock()
	calls = mock.calls.SetPayload
	mock.lockSetPayload.RUnlock()
	return calls
}

// SetResponse calls SetResponseFunc.
func (mock *EventMock) SetResponse(response string) {
	if mock.SetResponseFunc == nil {
		panic("EventMock.SetResponseFunc: method is nil but Event.SetResponse was just called")
	}
	callInfo := struct {
		Response string
	}{
		Response: response,
	}
	mock.lockSetResponse.Lock()
	mock.calls.SetResponse = append(mock.calls.SetResponse, callInfo)
	mock.lockSetResponse.Unlock()
	mock.SetResponseFunc(response)
}

// SetResponseCalls gets all the calls that were made to SetResponse.
// Check the length with:
//
//	len(mockedEvent.SetResponseCalls())
func (mock *EventMock) SetResponseCalls() []struct {
	Response string
} {
	var calls []struct {
		Response string
	}
	mock.lockSetResponse.RLock()
	calls = mock.calls.SetResponse
	mock.lockSetResponse.RUnlock()
	return calls
}

// Tenant calls TenantFunc.
func (mock *EventMock) Tenant() tenant.Id {
	if mock.TenantFunc == nil {
		panic("EventMock.TenantFunc: method is nil but Event.Tenant was just called")
	}
	callInfo := struct {
	}{}
	mock.lockTenant.Lock()
	mock.calls.Tenant = append(mock.calls.Tenant, callInfo)
	mock.lockTenant.Unlock()
	return mock.TenantFunc()
}

// TenantCalls gets all the calls that were made to Tenant.
// Check the length with:
//
//	len(mockedEvent.TenantCalls())
func (mock *EventMock) TenantCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockTenant.RLock()
	calls = mock.calls.Tenant
	mock.lockTenant.RUnlock()
	return calls
}

// UserTraceId calls UserTraceIdFunc.
func (mock *EventMock) UserTraceId() string {
	if mock.UserTraceIdFunc == nil {
		panic("EventMock.UserTraceIdFunc: method is nil but Event.UserTraceId was just called")
	}
	callInfo := struct {
	}{}
	mock.lockUserTraceId.Lock()
	mock.calls.UserTraceId = append(mock.calls.UserTraceId, callInfo)
	mock.lockUserTraceId.Unlock()
	return mock.UserTraceIdFunc()
}

// UserTraceIdCalls gets all the calls that were made to UserTraceId.
// Check the length with:
//
//	len(mockedEvent.UserTraceIdCalls())
func (mock *EventMock) UserTraceIdCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockUserTraceId.RLock()
	calls = mock.calls.UserTraceId
	mock.lockUserTraceId.RUnlock()
	return calls
}
