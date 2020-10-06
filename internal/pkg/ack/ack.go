package ack

import (
	"context"
	"sync"
)

// A SubTree is a subset of an ack Tree. Normally, users create SubTree when it expects to receive additional
// acknowledgements signaling the completions of the same work.
type SubTree interface {
	// Send an acknowledgement to indicate the completion of a work on this SubTree. Ack should only be called
	// after all NewSubTree(s) are created
	Ack()

	// Create a New SubTree for additional acknowledgements
	NewSubTree() (SubTree, error)
}

type Tree interface {
	SubTree

	// Block until either the completion function or the error function is called
	Wait()
}

// Construct a new ack Tree. The constructor takes in a context, a completion function, and an error function as its
// parameters. Once constructed, the ack Tree will start collecting acknowledgements asynchronously. When it receives
// all acknowledgements, it will call the completion function. If it does not receive all the acknowledgements before
// the context timeout/cancellation, it will call the error function. Once the completion function or the error function
// is called, the ack Tree is considered done, and subsequent acknowledgements are ignored.
func NewAckTree(ctx context.Context, fn func(), errFn func(error)) Tree {
	ack := &ackTree{
		parent:         nil,
		closure:        fn,
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

type InvalidActionError struct {
	errMsg string
}

func (e *InvalidActionError) Error() string {
	return e.errMsg
}

type TimeoutError struct {
	errMsg string
}

func (e *TimeoutError) Error() string {
	return e.errMsg
}

type ackTree struct {
	parent         *ackTree
	closure        func()
	errFn          func(error)
	numAckExpected int
	count          int
	lock           *sync.Mutex
	selfAcked      bool
	completed      bool
	wg             *sync.WaitGroup
}

func (ack *ackTree) Ack() {
	ack.ack(true)
}

func (ack *ackTree) NewSubTree() (SubTree, error) {
	ack.lock.Lock()
	defer ack.lock.Unlock()

	if ack.selfAcked {
		//cannot create more SubTree if you already acked yourself
		return nil, &InvalidActionError{"Already acked. Cannot create more subchain"}
	}

	ack.numAckExpected++
	return &ackTree{
		parent:         ack,
		closure:        nil,
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

func (ack *ackTree) ack(selfAck bool) {
	ack.lock.Lock()
	defer ack.lock.Unlock()

	if ack.completed {
		//this ack Tree is already done
		return
	}

	ack.count++
	if selfAck {
		ack.selfAcked = true
	}
	if ack.count == ack.numAckExpected {
		if ack.parent == nil && ack.closure != nil {
			ack.closure()
			ack.markComplete()
		} else {
			ack.parent.ack(false)
		}
	}
}

func (ack *ackTree) markComplete() {
	ack.completed = true
	ack.wg.Done()
}

func (ack *ackTree) startContextListener(ctx context.Context) {
	go func() {
		<-ctx.Done()
		ack.lock.Lock()
		defer ack.lock.Unlock()
		if ack.completed {
			//this ack Tree is already done
			return
		}
		ack.errFn(&TimeoutError{"Timeout reached"})
		ack.markComplete()
	}()
}
