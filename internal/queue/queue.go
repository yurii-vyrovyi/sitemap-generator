package queue

import (
	"io"
	"sync"
)

type (

	// ConcurrentQueue is a thread-safe FIFO queue.
	// Pop() returns a value is available or blocks until a new value will be pushed.
	// Pop() maybe called from many routines.
	// When Close() is called all blocked Pop() calls return io.EOF error.
	ConcurrentQueue struct {
		head *Element
		tail *Element

		cond *sync.Cond
		stop bool
	}

	Element struct {
		value    interface{}
		nextElem *Element
	}
)

func New() *ConcurrentQueue {
	return &ConcurrentQueue{
		cond: sync.NewCond(&sync.Mutex{}),
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

	newElem := Element{value: v}

	if q.tail == nil {
		q.head = &newElem
		q.tail = &newElem

	} else {
		q.tail.nextElem = &newElem
		q.tail = &newElem
	}

}

// Pop returns a value from the queue head.
// If the queue is empty Pop() waits until a new value will be pushed.
// After Close() call app blocked Pop() calls are released.
func (q *ConcurrentQueue) Pop() (interface{}, error) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for q.head == nil && !q.stop {
		q.cond.Wait()
	}

	if q.stop {
		return nil, io.EOF
	}

	head := q.head

	q.head = q.head.nextElem
	if q.head == nil {
		q.tail = nil
	}

	return head.value, nil
}
