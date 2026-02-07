package cache

import (
	"strconv"
	"sync"
	"testing"
)

func TestLRUBasic(t *testing.T) {
	l := NewLRU(2)

	l.Add("a", 1)
	l.Add("b", 2)

	if v, ok := l.Get("a"); !ok || v.(int) != 1 {
		t.Fatalf("expected a=1, got %#v, ok=%v", v, ok)
	}

	// adding c should evict b because a was recently used
	l.Add("c", 3)

	if _, ok := l.Get("b"); ok {
		t.Fatalf("expected b to be evicted")
	}

	if v, ok := l.Get("c"); !ok || v.(int) != 3 {
		t.Fatalf("expected c=3, got %#v, ok=%v", v, ok)
	}

	if v, ok := l.Peek("a"); !ok || v.(int) != 1 {
		t.Fatalf("expected peek a=1, got %#v, ok=%v", v, ok)
	}

	if !l.Remove("a") {
		t.Fatalf("expected a removed")
	}

	if _, ok := l.Get("a"); ok {
		t.Fatalf("a should not exist after remove")
	}

	l.Clear()
	if l.Len() != 0 {
		t.Fatalf("expected empty after Clear, len=%d", l.Len())
	}
}

func TestLRUConcurrency(t *testing.T) {
	l := NewLRU(100)
	var wg sync.WaitGroup

	// writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				key := "k" + strconv.Itoa((id+j)%50)
				l.Add(key, id*1000+j)
			}
		}(i)
	}

	// readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_ = l.Len()
			}
		}()
	}

	wg.Wait()

	// sanity: no panic, len <= capacity
	if l.Len() > 100 {
		t.Fatalf("len exceeds capacity: %d", l.Len())
	}
}
