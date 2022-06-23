package queue

import (
	"container/list"
	"io"
	"sync"
)

type (

	// ConcurrentQueue is a thread-safe FIFO queue.
	// Pop() returns a value is available or blocks until a new value will be pushed.
	// Pop() maybe called from many routines.
	// When Close() is called all blocked Pop() calls return io.EOF error.
	ConcurrentQueue struct {
		l *list.List

		cond *sync.Cond
		stop bool
	}
)

func New() *ConcurrentQueue {
	return &ConcurrentQueue{
		cond: sync.NewCond(&sync.Mutex{}),
		l:    list.New(),
	}
}

// Close unblocks all Pop calls.
// After Close() call Push() and Pop() calls don't affect a queue.
func (q *ConcurrentQueue) Close() {
	q.cond.L.Lock()

	q.stop = true

	q.cond.L.Unlock()
	q.cond.Broadcast()
}

// Push adds a value into the end of queue
func (q *ConcurrentQueue) Push(v interface{}) {

	q.cond.L.Lock()
	defer func() {
		q.cond.L.Unlock()
		q.cond.Signal()
	}()

	if q.stop {
		return
	}

	q.l.PushBack(v)
}

// Pop returns a value from the queue head.
// If the queue is empty Pop() waits until a new value will be pushed.
// After Close() call app blocked Pop() calls are released.
func (q *ConcurrentQueue) Pop() (interface{}, error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	var elem *list.Element

	for elem = q.l.Front(); elem == nil && !q.stop; elem = q.l.Front() {
		q.cond.Wait()
	}

	if q.stop {
		return nil, io.EOF
	}

	q.l.Remove(elem)

	return elem.Value, nil
}
