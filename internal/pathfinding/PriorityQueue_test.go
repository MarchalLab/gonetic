package pathfinding_test

import (
	"testing"

	"github.com/MarchalLab/gonetic/internal/pathfinding"
)

type Item struct {
	value int // The value of the item; arbitrary.
	index int // The index of the item in the heap; needed by update and is maintained by the methods of the heap.Interface.
}

// Priority implements queueItem.Priority.
// here we want the item with the lowest value to have the highest priority
func (i *Item) Priority(item *Item) bool {
	return i.value < item.value
}

// Index implements queueItem.Index.
func (i *Item) Index() int {
	return i.index
}

// SetIndex implements queueItem.SetIndex.
func (i *Item) SetIndex(index int) {
	i.index = index
}

func TestPush(t *testing.T) {
	pq := pathfinding.NewPriorityQueue[*Item]()
	item := &Item{value: 1}
	pq.Push(item)
	if pq.Len() != 1 {
		t.Errorf("Expected length 1, but got %d", pq.Len())
	}
}

func TestPriorityQueue(t *testing.T) {
	tests := []struct {
		name     string
		reverse  bool
		elements []*Item
		expected []int
	}{
		{
			name:     "test1",
			reverse:  false,
			elements: []*Item{{value: 3}, {value: 1}, {value: 2}},
			expected: []int{1, 2, 3},
		},
		{
			name:     "test2",
			reverse:  true,
			elements: []*Item{{value: 5}, {value: 1}, {value: 7}, {value: 3}},
			expected: []int{7, 5, 3, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pq *pathfinding.PriorityQueue[*Item]
			if tt.reverse {
				pq = pathfinding.NewReversePriorityQueue[*Item]()
			} else {
				pq = pathfinding.NewPriorityQueue[*Item]()
			}
			for i := range tt.elements {
				tt.elements[i].SetIndex(i)
				pq.Push(tt.elements[i])
			}

			for _, expected := range tt.expected {
				top := pq.Top()
				if top.value != expected {
					t.Errorf("Expected %v, but got top %v", expected, top)
				}
				item := pq.Pop()
				if item.value != expected {
					t.Errorf("Expected %v, but got pop %v", expected, item)
				}
			}
		})
	}
}
