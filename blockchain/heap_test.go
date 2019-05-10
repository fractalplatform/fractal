package blockchain

import "testing"

func TestMinHeap(t *testing.T) {
	mheap := &simpleHeap{
		cmp: func(a, b interface{}) int {
			return a.(int) - b.(int)
		},
	}
	for i := 0; i < 100; i++ {
		before := mheap.Len()
		mheap.push(i)
		if mheap.min().(int) != 0 {
			t.Fatal("top element not 0!")
		}
		if mheap.Len() != before+1 {
			t.Fatal("push failed!")
		}
	}
	for i := 0; i < 100; i++ {
		before := mheap.Len()
		if mheap.min().(int) != i {
			t.Fatal("top element is not min")
		}
		if mheap.pop().(int) != i {
			t.Fatal("pop is not min")
		}
		if mheap.Len() != before-1 {
			t.Fatal("push failed!")
		}
	}
}

func TestMaxHeap(t *testing.T) {
	mheap := &simpleHeap{
		cmp: func(a, b interface{}) int {
			return b.(int) - a.(int)
		},
	}
	for i := 0; i < 100; i++ {
		before := mheap.Len()
		mheap.push(i)
		if mheap.min().(int) != i {
			t.Fatal("top element not 0!")
		}
		if mheap.Len() != before+1 {
			t.Fatal("push failed!")
		}
	}
	for i := 100; i > 0; i-- {
		before := mheap.Len()
		if mheap.min().(int) != i-1 {
			t.Fatal("top element is not max")
		}
		if mheap.pop().(int) != i-1 {
			t.Fatal("pop is not min")
		}
		if mheap.Len() != before-1 {
			t.Fatal("push failed!")
		}
	}
}
