package ack_test

import (
	"context"
	"errors"
	ack "github.com/xmidt-org/ears/internal/pkg/ack"
	"testing"
	"time"
)

type TestWork struct {
	ack      ack.SubTree
	workTime time.Duration
}

func testWorker(ch chan *TestWork) {
	for work := range ch {
		time.Sleep(work.workTime)
		work.ack.Ack()
	}
}

func TestSingleAck(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second)

	workChan := make(chan *TestWork, 10)
	go testWorker(workChan)

	ack := ack.NewAckTree(ctx,
		func() {
			//we are good
		},
		func(err error) {
			t.Errorf("Receive error %s", err.Error())
		})

	//send a test work with an ack tree
	workChan <- &TestWork{ack: ack, workTime: 10 * time.Millisecond}
	ack.Wait()
}

func TestAckTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 100*time.Millisecond)

	workChan := make(chan *TestWork, 10)
	go testWorker(workChan)

	a := ack.NewAckTree(ctx,
		func() {
			t.Errorf("expecting timeout error. Completion function should not be called")
		},
		func(err error) {
			var expected *ack.TimeoutError
			if !errors.As(err, &expected) {
				t.Errorf("expecting timeout error, but get %s\n", err.Error())
			}
		})

	//send a test work with an ack tree
	workChan <- &TestWork{ack: a, workTime: 1 * time.Second}
	a.Wait()
}

func TestSubTree(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second)

	workChan := make(chan *TestWork, 10)
	for i := 0; i < 5; i++ {
		go testWorker(workChan)
	}

	a := ack.NewAckTree(ctx,
		func() {
			//we are good
		},
		func(err error) {
			t.Errorf("Receive error %s", err.Error())
		})

	//send 100 test work with an ack tree
	for i := 0; i < 100; i++ {
		subTree, err := a.NewSubTree()
		if err != nil {
			t.Fatalf("Fail to create subtree %s\n", err.Error())
		}
		workChan <- &TestWork{ack: subTree, workTime: 10 * time.Millisecond}
	}
	a.Ack()
	a.Wait()

	var invalidErr *ack.InvalidActionError

	//Test error case. Should not be able to create subtree anymorea
	_, err := a.NewSubTree()
	if err == nil || !errors.As(err, &invalidErr) {
		t.Fatalf("Expect INvalidActionError, but got %v\n", err)
	}
}

func level1Worker(t *testing.T, ch chan *TestWork, ch2 chan *TestWork) {
	for work := range ch {
		time.Sleep(work.workTime)
		//split 10 ways
		for i := 0; i < 10; i++ {
			subTree, err := work.ack.NewSubTree()
			if err != nil {
				t.Fatalf("Fail to create subtree %s\n", err.Error())
			}
			ch2 <- &TestWork{ack: subTree, workTime: 0 * time.Millisecond}
		}
		work.ack.Ack()
	}
}

func Test2LevelSubTree(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second)

	workChan := make(chan *TestWork, 100)
	workChan2 := make(chan *TestWork, 1000)
	for i := 0; i < 5; i++ {
		go level1Worker(t, workChan, workChan2)
	}
	for i := 0; i < 10; i++ {
		go testWorker(workChan2)
	}

	ack := ack.NewAckTree(ctx,
		func() {
			//we are good
		},
		func(err error) {
			t.Errorf("Receive error %s", err.Error())
		})

	//send 100 test work with an ack tree
	for i := 0; i < 100; i++ {
		subTree, err := ack.NewSubTree()
		if err != nil {
			t.Fatalf("Fail to create subtree %s\n", err.Error())
		}
		workChan <- &TestWork{ack: subTree, workTime: 1 * time.Millisecond}
	}
	ack.Ack()
	ack.Wait()
}
