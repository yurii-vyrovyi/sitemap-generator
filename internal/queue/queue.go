package queue

import (
	"io"
	"sync"
)

type (
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

func (q *ConcurrentQueue) Close() {
	q.cond.L.Lock()

	q.stop = true

	q.cond.L.Unlock()
	q.cond.Broadcast()
}

func (q *ConcurrentQueue) Push(v interface{}) {

	q.cond.L.Lock()

	newElem := Element{value: v}

	if q.tail == nil {
		q.head = &newElem
		q.tail = &newElem

	} else {
		q.tail.nextElem = &newElem
		q.tail = &newElem
	}

	q.cond.L.Unlock()
	q.cond.Signal()
}

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
