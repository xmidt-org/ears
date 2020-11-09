// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package sender

import (
	"context"
	"github.com/xmidt-org/ears/pkg/event"
	"sync"
)

// Ensure, that HasherMock does implement Hasher.
// If this is not the case, regenerate this file with moq.
var _ Hasher = &HasherMock{}

// HasherMock is a mock implementation of Hasher.
//
//     func TestSomethingThatUsesHasher(t *testing.T) {
//
//         // make and configure a mocked Hasher
//         mockedHasher := &HasherMock{
//             SenderHashFunc: func(config string) (string, error) {
// 	               panic("mock out the SenderHash method")
//             },
//         }
//
//         // use mockedHasher in code that requires Hasher
//         // and then make assertions.
//
//     }
type HasherMock struct {
	// SenderHashFunc mocks the SenderHash method.
	SenderHashFunc func(config string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// SenderHash holds details about calls to the SenderHash method.
		SenderHash []struct {
			// Config is the config argument value.
			Config string
		}
	}
	lockSenderHash sync.RWMutex
}

// SenderHash calls SenderHashFunc.
func (mock *HasherMock) SenderHash(config string) (string, error) {
	if mock.SenderHashFunc == nil {
		panic("HasherMock.SenderHashFunc: method is nil but Hasher.SenderHash was just called")
	}
	callInfo := struct {
		Config string
	}{
		Config: config,
	}
	mock.lockSenderHash.Lock()
	mock.calls.SenderHash = append(mock.calls.SenderHash, callInfo)
	mock.lockSenderHash.Unlock()
	return mock.SenderHashFunc(config)
}

// SenderHashCalls gets all the calls that were made to SenderHash.
// Check the length with:
//     len(mockedHasher.SenderHashCalls())
func (mock *HasherMock) SenderHashCalls() []struct {
	Config string
} {
	var calls []struct {
		Config string
	}
	mock.lockSenderHash.RLock()
	calls = mock.calls.SenderHash
	mock.lockSenderHash.RUnlock()
	return calls
}

// Ensure, that NewSendererMock does implement NewSenderer.
// If this is not the case, regenerate this file with moq.
var _ NewSenderer = &NewSendererMock{}

// NewSendererMock is a mock implementation of NewSenderer.
//
//     func TestSomethingThatUsesNewSenderer(t *testing.T) {
//
//         // make and configure a mocked NewSenderer
//         mockedNewSenderer := &NewSendererMock{
//             NewSenderFunc: func(config string) (Sender, error) {
// 	               panic("mock out the NewSender method")
//             },
//             SenderHashFunc: func(config string) (string, error) {
// 	               panic("mock out the SenderHash method")
//             },
//         }
//
//         // use mockedNewSenderer in code that requires NewSenderer
//         // and then make assertions.
//
//     }
type NewSendererMock struct {
	// NewSenderFunc mocks the NewSender method.
	NewSenderFunc func(config string) (Sender, error)

	// SenderHashFunc mocks the SenderHash method.
	SenderHashFunc func(config string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// NewSender holds details about calls to the NewSender method.
		NewSender []struct {
			// Config is the config argument value.
			Config string
		}
		// SenderHash holds details about calls to the SenderHash method.
		SenderHash []struct {
			// Config is the config argument value.
			Config string
		}
	}
	lockNewSender  sync.RWMutex
	lockSenderHash sync.RWMutex
}

// NewSender calls NewSenderFunc.
func (mock *NewSendererMock) NewSender(config string) (Sender, error) {
	if mock.NewSenderFunc == nil {
		panic("NewSendererMock.NewSenderFunc: method is nil but NewSenderer.NewSender was just called")
	}
	callInfo := struct {
		Config string
	}{
		Config: config,
	}
	mock.lockNewSender.Lock()
	mock.calls.NewSender = append(mock.calls.NewSender, callInfo)
	mock.lockNewSender.Unlock()
	return mock.NewSenderFunc(config)
}

// NewSenderCalls gets all the calls that were made to NewSender.
// Check the length with:
//     len(mockedNewSenderer.NewSenderCalls())
func (mock *NewSendererMock) NewSenderCalls() []struct {
	Config string
} {
	var calls []struct {
		Config string
	}
	mock.lockNewSender.RLock()
	calls = mock.calls.NewSender
	mock.lockNewSender.RUnlock()
	return calls
}

// SenderHash calls SenderHashFunc.
func (mock *NewSendererMock) SenderHash(config string) (string, error) {
	if mock.SenderHashFunc == nil {
		panic("NewSendererMock.SenderHashFunc: method is nil but NewSenderer.SenderHash was just called")
	}
	callInfo := struct {
		Config string
	}{
		Config: config,
	}
	mock.lockSenderHash.Lock()
	mock.calls.SenderHash = append(mock.calls.SenderHash, callInfo)
	mock.lockSenderHash.Unlock()
	return mock.SenderHashFunc(config)
}

// SenderHashCalls gets all the calls that were made to SenderHash.
// Check the length with:
//     len(mockedNewSenderer.SenderHashCalls())
func (mock *NewSendererMock) SenderHashCalls() []struct {
	Config string
} {
	var calls []struct {
		Config string
	}
	mock.lockSenderHash.RLock()
	calls = mock.calls.SenderHash
	mock.lockSenderHash.RUnlock()
	return calls
}

// Ensure, that SenderMock does implement Sender.
// If this is not the case, regenerate this file with moq.
var _ Sender = &SenderMock{}

// SenderMock is a mock implementation of Sender.
//
//     func TestSomethingThatUsesSender(t *testing.T) {
//
//         // make and configure a mocked Sender
//         mockedSender := &SenderMock{
//             SendFunc: func(ctx context.Context, e event.Event) error {
// 	               panic("mock out the Send method")
//             },
//         }
//
//         // use mockedSender in code that requires Sender
//         // and then make assertions.
//
//     }
type SenderMock struct {
	// SendFunc mocks the Send method.
	SendFunc func(ctx context.Context, e event.Event) error

	// calls tracks calls to the methods.
	calls struct {
		// Send holds details about calls to the Send method.
		Send []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// E is the e argument value.
			E event.Event
		}
	}
	lockSend sync.RWMutex
}

// Send calls SendFunc.
func (mock *SenderMock) Send(ctx context.Context, e event.Event) error {
	if mock.SendFunc == nil {
		panic("SenderMock.SendFunc: method is nil but Sender.Send was just called")
	}
	callInfo := struct {
		Ctx context.Context
		E   event.Event
	}{
		Ctx: ctx,
		E:   e,
	}
	mock.lockSend.Lock()
	mock.calls.Send = append(mock.calls.Send, callInfo)
	mock.lockSend.Unlock()
	return mock.SendFunc(ctx, e)
}

// SendCalls gets all the calls that were made to Send.
// Check the length with:
//     len(mockedSender.SendCalls())
func (mock *SenderMock) SendCalls() []struct {
	Ctx context.Context
	E   event.Event
} {
	var calls []struct {
		Ctx context.Context
		E   event.Event
	}
	mock.lockSend.RLock()
	calls = mock.calls.Send
	mock.lockSend.RUnlock()
	return calls
}
