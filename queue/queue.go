package queue

import (
	"container/heap"
	"sync"
	"time"

	"github.com/dyrkin/rezerwacje-duw-go/config"
)

type Reservation struct {
	Entity   *config.Entity
	Date     string
	Term     string
	UserData *[]*config.Row
}

type item struct {
	reservation *Reservation
	priority    int64
	index       int
}

type priorityQueue []*item

type ReservationQueue struct {
	pq           *priorityQueue
	items        map[Reservation]bool
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
	items := map[Reservation]bool{}
	heap.Init(pq)
	lock := &sync.Mutex{}
	pushLock := &sync.Mutex{}
	nonEmptyCond := sync.NewCond(lock)
	return &ReservationQueue{pq, items, time.Now(), lock, pushLock, nonEmptyCond, -1}
}

func NewWithLimit(limit int) *ReservationQueue {
	q := New()
	q.limit = limit
	return q
}

func (pq *priorityQueue) fix(i int) {
	heap.Fix(pq, i)
}

func (q *ReservationQueue) update(reservation *Reservation, i int) {
	item := (*q.pq)[i]
	item.reservation = reservation
	item.priority = int64(time.Now().Sub(q.time))
	q.pq.fix(i)
}

func (q *ReservationQueue) push(item *item) {
	heap.Push(q.pq, item)
}

//TODO Find more optimal way how to find index of item with lowest priority
func (pq *priorityQueue) Lowest() int {
	var item *item
	for _, v := range *pq {
		if (item == nil) || (v.priority < item.priority) {
			item = v
		}
	}
	return item.index
}

func (pq *priorityQueue) Index(reservation Reservation) int {
	for _, v := range *pq {
		if *v.reservation == reservation {
			return v.index
		}
	}
	return -1
}

func (q *ReservationQueue) Push(reservation *Reservation) {
	q.pushLock.Lock()
	defer q.pushLock.Unlock()
	if _, ok := q.items[*reservation]; ok {
		q.update(reservation, q.pq.Index(*reservation))
	} else {
		if q.limit != -1 && q.len() == q.limit {
			q.update(reservation, q.pq.Lowest())
		} else {
			item := &item{reservation: reservation, priority: int64(time.Now().Sub(q.time))}
			q.push(item)
		}
	}
	q.items[*reservation] = true
	q.nonEmptyCond.Signal()
}

func (q *ReservationQueue) pop() *Reservation {
	if q.pq.Len() > 0 {
		reservation := heap.Pop(q.pq).(*item).reservation
		delete(q.items, *reservation)
		return reservation
	}
	return nil
}

func (q *ReservationQueue) Pop() *Reservation {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.pop()
}

func (q *ReservationQueue) Take() *Reservation {
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
