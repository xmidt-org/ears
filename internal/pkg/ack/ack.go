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

package ack

import (
	"context"
	"github.com/xmidt-org/ears/pkg/errs"
	"sync"
)

// A SubTree is a subset of an ack Tree. Normally, users create SubTree when it expects to receive additional
// acknowledgements signaling the completions of the same work.
type SubTree interface {
	// Send an acknowledgement to indicate the completion of a work on this SubTree. Ack should only be called
	// after all NewSubTree(s) are created
	Ack()

	// Send an negative acknowledgement to indicate that there is an error completing a work on this SubTree.
	// Nack error will be wrapping in a NackError and sent to the error function.
	Nack(err error)

	// Create a New SubTree for additional acknowledgements
	NewSubTree() (SubTree, error)

	// Whether or not this node is acked or nacked (not necessarily the whole tree)
	IsAcked() bool
}

type Tree interface {
	SubTree

	// Block until either the completion function or the error function is called
	Wait()

	// Whether or not the tree is completed
	IsDone() bool
}

// Construct a new ack Tree. The constructor takes in a context, a completion function, and an error function as its
// parameters. Once constructed, the ack Tree will start collecting acknowledgements asynchronously. When it receives
// all acknowledgements, it will call the completion function. If it does not receive all the acknowledgements before
// the context timeout/cancellation, it will call the error function. Once the completion function or the error function
// is called, the ack Tree is considered done, and subsequent acknowledgements are ignored.
func NewAckTree(ctx context.Context, fn func(), errFn func(error)) Tree {
	ack := &ackTree{
		parent:         nil,
		completedFn:    fn,
		errFn:          errFn,
		numAckExpected: 1,
		count:          0,
		lock:           &sync.Mutex{},
		selfAcked:      false,
		wg:             &sync.WaitGroup{},
	}
	ack.wg.Add(1)
	ack.startContextListener(ctx)
	return ack
}

type AlreadyAckedError struct {
}

func (e *AlreadyAckedError) Error() string {
	return errs.String("AlreadyAckedError", nil, nil)
}

type TimeoutError struct {
}

func (e *TimeoutError) Error() string {
	return errs.String("TimeoutError", nil, nil)
}

type NackError struct {
	err error
}

func (e *NackError) Error() string {
	return errs.String("NackError", nil, e.err)
}

func (e *NackError) Unwrap() error {
	return e.err
}

type ackTree struct {
	parent         *ackTree
	completedFn    func()
	errFn          func(error)
	numAckExpected int
	count          int
	lock           *sync.Mutex
	selfAcked      bool
	completed      bool
	wg             *sync.WaitGroup
}

func (ack *ackTree) Ack() {
	ack.ack(true, nil)
}

func (ack *ackTree) Nack(err error) {
	ack.ack(true, &NackError{err})
}

func (ack *ackTree) NewSubTree() (SubTree, error) {
	ack.lock.Lock()
	defer ack.lock.Unlock()

	if ack.selfAcked {
		//cannot create more SubTree if you already acked yourself
		return nil, &AlreadyAckedError{}
	}

	ack.numAckExpected++
	return &ackTree{
		parent:         ack,
		completedFn:    nil,
		errFn:          nil,
		numAckExpected: 1,
		count:          0,
		lock:           &sync.Mutex{},
		selfAcked:      false,
	}, nil
}

//wait until closure functions are called
func (ack *ackTree) Wait() {
	ack.wg.Wait()
}

func (ack *ackTree) ack(selfAck bool, err error) {
	ack.lock.Lock()
	defer ack.lock.Unlock()

	if ack.completed {
		//this ack Tree is already done
		return
	}

	if err != nil {
		//there is an error, propagate up to parent
		if ack.parent == nil && ack.errFn != nil {
			go ack.errFn(err)
			ack.markComplete()
		} else {
			ack.parent.ack(false, err)
		}
		return
	}

	ack.count++
	if selfAck {
		ack.selfAcked = true
	}
	if ack.count == ack.numAckExpected {
		if ack.parent == nil && ack.completedFn != nil {
			go ack.completedFn()
			ack.markComplete()
		} else {
			ack.parent.ack(false, nil)
			ack.completed = true
		}
	}
}

func (ack *ackTree) markComplete() {
	ack.completed = true
	ack.wg.Done()
}

func (ack *ackTree) startContextListener(ctx context.Context) {
	go func() {
		done := ctx.Done()
		if done == nil {
			//this context can never be canceled. No need to listen
			return
		}
		<-done
		ack.lock.Lock()
		defer ack.lock.Unlock()
		if ack.completed {
			//this ack Tree is already done
			return
		}
		go ack.errFn(&TimeoutError{})
		ack.markComplete()
	}()
}

func (ack *ackTree) IsDone() bool {
	ack.lock.Lock()
	defer ack.lock.Unlock()

	return ack.completed
}

func (ack *ackTree) IsAcked() bool {
	ack.lock.Lock()
	defer ack.lock.Unlock()

	return ack.selfAcked
}
