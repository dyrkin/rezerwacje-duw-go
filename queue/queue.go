package queue

import (
	"container/heap"
	"sync"
	"time"
)

type item struct {
	reservationFn ReservationFn
	priority      time.Duration
	index         int
}

type priorityQueue []*item

type ReservationFn func()

type ReservationQueue struct {
	pq           *priorityQueue
	time         time.Time
	lock         *sync.Mutex
	nonEmptyCond *sync.Cond
}

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func New() *ReservationQueue {
	pq := &priorityQueue{}
	heap.Init(pq)
	lock := &sync.Mutex{}
	nonEmptyCond := sync.NewCond(lock)
	return &ReservationQueue{pq, time.Now(), lock, nonEmptyCond}
}

func (q *ReservationQueue) Push(fn ReservationFn) {
	item := &item{reservationFn: fn, priority: time.Now().Sub(q.time)}
	heap.Push(q.pq, item)
	q.nonEmptyCond.Signal()
}

func (q *ReservationQueue) pop() ReservationFn {
	if q.pq.Len() > 0 {
		return heap.Pop(q.pq).(*item).reservationFn
	}
	return nil
}

func (q *ReservationQueue) Pop() ReservationFn {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pop()
}

func (q *ReservationQueue) Take() ReservationFn {
	q.lock.Lock()
	defer q.lock.Unlock()
	for q.pq.Len() == 0 {
		q.nonEmptyCond.Wait()
	}
	return q.pop()
}

func (q *ReservationQueue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pq.Len()
}
