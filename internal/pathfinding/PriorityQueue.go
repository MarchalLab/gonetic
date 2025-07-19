package pathfinding

import (
	"container/heap"
)

type queueItem[T any] interface {
	Priority(other T) bool
	Index() int
	SetIndex(int)
}

// queue is a priority queue of items
type queue[T queueItem[T]] struct {
	items   []T
	reverse bool
}

// Len is part of sort.Interface
func (q queue[T]) Len() int {
	return len(q.items)
}

// Less is part of sort.Interface
// whether the element with index i must sort before the element with index j.
// when q.reverse is true, the element with index j must sort before the element with index i.
func (q queue[T]) Less(i, j int) bool {
	if q.reverse {
		return q.items[j].Priority(q.items[i])
	}
	return q.items[i].Priority(q.items[j])
}

// Swap is part of sort.Interface
func (q queue[T]) Swap(i, j int) {
	q.items[i], q.items[j] = q.items[j], q.items[i]
	q.items[i].SetIndex(i)
	q.items[j].SetIndex(j)
}

// Push is part of heap.Interface
// add x as element Len()
func (q *queue[T]) Push(x interface{}) {
	(*q).items = append((*q).items, x.(T))
}

// Pop is part of heap.Interface
// remove and return element Len() - 1.
func (q *queue[T]) Pop() interface{} {
	old := (*q).items
	n := len(old)
	item := old[n-1]
	(*q).items = old[0 : n-1]
	return item
}

// PriorityQueue is a public priority queue using a queue internally
type PriorityQueue[T queueItem[T]] struct {
	queue *queue[T]
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue[T queueItem[T]]() *PriorityQueue[T] {
	q := &queue[T]{reverse: false}
	heap.Init(q)
	return &PriorityQueue[T]{
		queue: q,
	}
}

// NewReversePriorityQueue creates a new priority queue that is reversed
func NewReversePriorityQueue[T queueItem[T]]() *PriorityQueue[T] {
	q := &queue[T]{reverse: true}
	heap.Init(q)
	return &PriorityQueue[T]{
		queue: q,
	}
}

// Push adds an item to the priority queue
func (pq *PriorityQueue[T]) Push(item T) {
	heap.Push(pq.queue, item)
}

// Pop removes the top item from the priority queue
func (pq *PriorityQueue[T]) Pop() T {
	return heap.Pop(pq.queue).(T)
}

// Len returns the length of the priority queue
func (pq *PriorityQueue[T]) Len() int { return pq.queue.Len() }

// Empty returns true if the priority queue is empty
func (pq *PriorityQueue[T]) Empty() bool { return pq.Len() == 0 }

// Top returns the top item of the priority queue
func (pq *PriorityQueue[T]) Top() T {
	return (*pq.queue).items[0]
}

func (pq *PriorityQueue[T]) PopToReverseSlice() []T {
	slice := make([]T, pq.Len())
	idx := len(slice)
	for !pq.Empty() {
		idx--
		slice[idx] = heap.Pop(pq.queue).(T)
	}
	return slice
}

func (pq *PriorityQueue[T]) TrimBelow(score func(T) float64, cutoff float64) int {
	newItems := make([]T, 0, len(pq.queue.items))
	trimmed := 0
	for _, item := range pq.queue.items {
		if score(item) > cutoff {
			newItems = append(newItems, item)
		}
		trimmed++
	}
	pq.queue.items = newItems
	// rebuild heap
	heap.Init(pq.queue)
	return trimmed
}
