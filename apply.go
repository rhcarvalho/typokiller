package typokiller

import (
	"container/heap"
	"fmt"
	"io/ioutil"
)

// Apply replaces misspelled words with their respective replacements.
// It processes changes from bottom to the top of files to do not invalidate
// offsets.
func Apply(misspellings []*Misspelling, status chan string) {
	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(misspellings))
	for i, m := range misspellings {
		pq[i] = &Item{
			value:    m,
			priority: m.Text.Position.Offset + m.Offset,
			index:    i,
		}
	}
	heap.Init(&pq)

	// Take the items out; they arrive in decreasing priority order.
	for max := len(pq); max > 0; max-- {
		item := heap.Pop(&pq).(*Item)
		m := item.value
		if m.Action.Type == Replace {
			status <- "."
			pos := m.Text.Position

			b, err := ioutil.ReadFile(pos.Filename)
			if err != nil {
				panic(err)
			}
			begin := pos.Offset + m.Offset
			end := begin + len(m.Word)
			found := string(b[begin:end])
			if found == m.Word {
				replaced := replaceSlice(b, begin, end, []byte(m.Action.Replacement)...)
				ioutil.WriteFile(pos.Filename, replaced, 0644)
			} else {
				status <- fmt.Sprintf("(%s != %s)", found, m.Word)
			}
		}
	}
	close(status)
}

// replaceSlice replaces part of a byte slice with a byte or slice.
// This is similar in intent to slice assignment as implemented in Python:
//   a[3:6] = b[1:4]
func replaceSlice(slice []byte, begin, end int, repl ...byte) []byte {
	total := len(slice) - (end - begin) + len(repl)
	if total > cap(slice) {
		newSlice := make([]byte, total)
		copy(newSlice, slice)
		slice = newSlice
	}
	originalLen := len(slice)
	slice = slice[:total]
	copy(slice[begin+len(repl):originalLen], slice[end:originalLen])
	copy(slice[begin:begin+len(repl)], repl)
	return slice
}

// An Item is something we manage in a priority queue.
type Item struct {
	value    *Misspelling // The value of the item; arbitrary.
	priority int          // The priority of the item in the queue.
	index    int          // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
