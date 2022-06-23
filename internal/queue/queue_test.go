package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueue_OneThread(t *testing.T) {

	q := New()

	for i := 0; i < 10; i++ {
		q.Push(fmt.Sprintf("val-%d", i))
	}

	for i := 0; i < 10; i++ {
		v, err := q.Pop()
		if err != nil {
			return
		}

		fmt.Println(v)
	}

}

func TestQueue_Parallel(t *testing.T) {

	q := New()

	go func() {
		for {
			v, err := q.Pop()
			if err != nil {
				return
			}

			s := v.(string)
			fmt.Println(s)
		}
	}()

	time.Sleep(1 * time.Second)

	for i := 0; i < 10; i++ {
		q.Push(fmt.Sprintf("val-%d", i))
	}

	time.Sleep(1 * time.Second)
	q.Close()

}

func Test_Queue(t *testing.T) {

	q := New()

	const (
		NRoutines      = 5
		NRoutinePoints = 200
	)
	wg := sync.WaitGroup{}

	mapRes := make(map[string]int)

	muxRes := sync.Mutex{}

	wgPop := sync.WaitGroup{}
	wgPop.Add(1)
	go func() {
		defer func() {
			wgPop.Done()
			fmt.Println("wgPop.Done()")
		}()

		for {

			v, err := q.Pop()
			if err != nil {
				return
			}

			strV, _ := v.(string)

			muxRes.Lock()
			mapRes[strV] = mapRes[strV] + 1
			muxRes.Unlock()
		}
	}()

	for iRoutine := 0; iRoutine < NRoutines; iRoutine++ {

		wg.Add(1)
		go func(iRoutine int) {
			defer wg.Done()

			for iPoint := 0; iPoint < NRoutinePoints; iPoint++ {
				key := fmt.Sprintf("%d:%d", iRoutine, iPoint)
				q.Push(key)
			}

		}(iRoutine)
	}

	wg.Wait()

	// a delay to allow Pop worker to drain the queue
	time.Sleep(1 * time.Second)

	q.Close()
	wgPop.Wait()

	require.Equal(t, NRoutines*NRoutinePoints, len(mapRes))

	for iRoutine := 0; iRoutine < NRoutines; iRoutine++ {
		for iPoint := 0; iPoint < NRoutinePoints; iPoint++ {
			key := fmt.Sprintf("%d:%d", iRoutine, iPoint)

			v, ok := mapRes[key]
			require.True(t, ok)
			require.Equal(t, v, 1)
		}
	}

}
