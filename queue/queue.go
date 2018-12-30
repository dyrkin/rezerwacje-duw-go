package queue

import (
	"container/heap"
	"sync"
	"time"
)

type item struct {
	reservationFn ReservationFn
	priority      int64
	index         int
}

type priorityQueue []*item

type ReservationFn func()

type ReservationQueue struct {
	pq           *priorityQueue
	time         time.Time
	lock         *sync.Mutex
	pushLock     *sync.Mutex
	nonEmptyCond *sync.Cond
	limit        int
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
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func New() *ReservationQueue {
	pq := &priorityQueue{}
	heap.Init(pq)
	lock := &sync.Mutex{}
	pushLock := &sync.Mutex{}
	nonEmptyCond := sync.NewCond(lock)
	return &ReservationQueue{pq, time.Now(), lock, pushLock, nonEmptyCond, -1}
}

func NewWithLimit(limit int) *ReservationQueue {
	q := New()
	q.limit = limit
	return q
}

func (pq *priorityQueue) fix(i int) {
	heap.Fix(pq, i)
}

func (q *ReservationQueue) update(fn ReservationFn, i int) {
	item := (*q.pq)[i]
	item.reservationFn = fn
	item.priority = int64(time.Now().Sub(q.time))
	q.pq.fix(i)
}

func (q *ReservationQueue) push(item *item) {
	heap.Push(q.pq, item)
}

func (pq *priorityQueue) Lowest() int {
	var item *item
	for _, v := range *pq {
		if (item == nil) || (v.priority < item.priority) {
			item = v
		}
	}
	return item.index
}

func (q *ReservationQueue) Push(fn ReservationFn) {
	q.pushLock.Lock()
	defer q.pushLock.Unlock()
	if q.limit != -1 && q.len() == q.limit {
		q.update(fn, q.pq.Lowest())
	} else {
		item := &item{reservationFn: fn, priority: int64(time.Now().Sub(q.time))}
		q.push(item)
	}
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

func (q *ReservationQueue) len() int {
	return q.pq.Len()
}

func (q *ReservationQueue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.len()
}
