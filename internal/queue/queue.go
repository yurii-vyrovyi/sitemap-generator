package queue

import (
	"container/list"
	"io"
	"sync"
	"sync/atomic"
)

type (

	// ConcurrentQueue is a thread-safe FIFO queue.
	// Pop() returns a value is available or blocks until a new value will be pushed.
	// Pop() maybe called from many routines.
	// When Close() is called all blocked Pop() calls return io.EOF error.
	ConcurrentQueue struct {
		l *list.List

		cond *sync.Cond
		stop atomic.Value
	}
)

func New() *ConcurrentQueue {
	q := ConcurrentQueue{
		cond: sync.NewCond(&sync.Mutex{}),
		l:    list.New(),
	}

	q.stop.Store(false)

	return &q
}

// Close unblocks all Pop calls.
// After Close() call Push() and Pop() calls don't affect a queue.
func (q *ConcurrentQueue) Close() {
	q.cond.L.Lock()

	q.stop.Store(true)

	q.cond.L.Unlock()
	q.cond.Broadcast()
}

// Push adds a value into the end of queue
func (q *ConcurrentQueue) Push(v interface{}) {
	if q.stop.Load().(bool) {
		return
	}

	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()

	q.l.PushBack(v)
}

// Pop returns a value from the queue head.
// If the queue is empty Pop() waits until a new value will be pushed.
// After Close() call app blocked Pop() calls are released.
func (q *ConcurrentQueue) Pop() (interface{}, error) {
	if q.stop.Load().(bool) {
		return nil, io.EOF
	}

	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	var elem *list.Element

	for elem = q.l.Front(); elem == nil && !q.stop.Load().(bool); elem = q.l.Front() {
		q.cond.Wait()
	}

	if q.stop.Load().(bool) {
		return nil, io.EOF
	}

	q.l.Remove(elem)

	return elem.Value, nil
}
