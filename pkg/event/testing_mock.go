// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package event

import (
	"context"
	"sync"
)

// Ensure, that EventMock does implement Event.
// If this is not the case, regenerate this file with moq.
var _ Event = &EventMock{}

// EventMock is a mock implementation of Event.
//
// 	func TestSomethingThatUsesEvent(t *testing.T) {
//
// 		// make and configure a mocked Event
// 		mockedEvent := &EventMock{
// 			AckFunc: func()  {
// 				panic("mock out the Ack method")
// 			},
// 			CloneFunc: func(ctx context.Context) (Event, error) {
// 				panic("mock out the Clone method")
// 			},
// 			ContextFunc: func() context.Context {
// 				panic("mock out the Context method")
// 			},
// 			MetadataFunc: func() interface{} {
// 				panic("mock out the Metadata method")
// 			},
// 			NackFunc: func(err error)  {
// 				panic("mock out the Nack method")
// 			},
// 			PayloadFunc: func() interface{} {
// 				panic("mock out the Payload method")
// 			},
// 			SetContextFunc: func(ctx context.Context) error {
// 				panic("mock out the SetContext method")
// 			},
// 			SetMetadataFunc: func(metadata interface{}) error {
// 				panic("mock out the SetMetadata method")
// 			},
// 			SetPayloadFunc: func(payload interface{}) error {
// 				panic("mock out the SetPayload method")
// 			},
// 		}
//
// 		// use mockedEvent in code that requires Event
// 		// and then make assertions.
//
// 	}
type EventMock struct {
	// AckFunc mocks the Ack method.
	AckFunc func()

	// CloneFunc mocks the Clone method.
	CloneFunc func(ctx context.Context) (Event, error)

	// ContextFunc mocks the Context method.
	ContextFunc func() context.Context

	// MetadataFunc mocks the Metadata method.
	MetadataFunc func() interface{}

	// NackFunc mocks the Nack method.
	NackFunc func(err error)

	// PayloadFunc mocks the Payload method.
	PayloadFunc func() interface{}

	// SetContextFunc mocks the SetContext method.
	SetContextFunc func(ctx context.Context) error

	// SetMetadataFunc mocks the SetMetadata method.
	SetMetadataFunc func(metadata interface{}) error

	// SetPayloadFunc mocks the SetPayload method.
	SetPayloadFunc func(payload interface{}) error

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
		// SetContext holds details about calls to the SetContext method.
		SetContext []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// SetMetadata holds details about calls to the SetMetadata method.
		SetMetadata []struct {
			// Metadata is the metadata argument value.
			Metadata interface{}
		}
		// SetPayload holds details about calls to the SetPayload method.
		SetPayload []struct {
			// Payload is the payload argument value.
			Payload interface{}
		}
	}
	lockAck         sync.RWMutex
	lockClone       sync.RWMutex
	lockContext     sync.RWMutex
	lockMetadata    sync.RWMutex
	lockNack        sync.RWMutex
	lockPayload     sync.RWMutex
	lockSetContext  sync.RWMutex
	lockSetMetadata sync.RWMutex
	lockSetPayload  sync.RWMutex
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
//     len(mockedEvent.AckCalls())
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
//     len(mockedEvent.CloneCalls())
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
//     len(mockedEvent.ContextCalls())
func (mock *EventMock) ContextCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockContext.RLock()
	calls = mock.calls.Context
	mock.lockContext.RUnlock()
	return calls
}

// Metadata calls MetadataFunc.
func (mock *EventMock) Metadata() interface{} {
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
//     len(mockedEvent.MetadataCalls())
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
//     len(mockedEvent.NackCalls())
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
//     len(mockedEvent.PayloadCalls())
func (mock *EventMock) PayloadCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPayload.RLock()
	calls = mock.calls.Payload
	mock.lockPayload.RUnlock()
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
//     len(mockedEvent.SetContextCalls())
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
func (mock *EventMock) SetMetadata(metadata interface{}) error {
	if mock.SetMetadataFunc == nil {
		panic("EventMock.SetMetadataFunc: method is nil but Event.SetMetadata was just called")
	}
	callInfo := struct {
		Metadata interface{}
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
//     len(mockedEvent.SetMetadataCalls())
func (mock *EventMock) SetMetadataCalls() []struct {
	Metadata interface{}
} {
	var calls []struct {
		Metadata interface{}
	}
	mock.lockSetMetadata.RLock()
	calls = mock.calls.SetMetadata
	mock.lockSetMetadata.RUnlock()
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
//     len(mockedEvent.SetPayloadCalls())
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
