package cache

import (
	"testing"
)

type fakeEviction struct {
	Values map[interface{}]int
	Log    Printf
}

func (e *fakeEviction) Added(key interface{}) {
	e.Log("Added %T(%v)", key, key)
	e.Values[key] = 0
}

func (e *fakeEviction) Removed(key interface{}) (removed bool) {
	if _, removed = e.Values[key]; removed {
		delete(e.Values, key)
	}
	e.Log("Removed %T(%v) -> %v", key, key, removed)
	return
}

func (e *fakeEviction) Hit(key interface{}) {
	e.Log("Hit %T(%v)", key, key)
	e.Values[key]++
}

func (e *fakeEviction) Pop() (key interface{}) {
	min := 1000
	for k, n := range e.Values {
		if key == nil || n < min {
			key, min = k, n
		}
	}
	if key != nil {
		delete(e.Values, key)
	}
	e.Log("Pop -> %T(%v)", key, key)
	return
}

func TestEvictingCache(t *testing.T) {

	e := &fakeEviction{make(map[interface{}]int), t.Logf}

	c := NewMemoryStorage(Spy(t.Logf), Eviction(3, e), Spy(t.Logf))

	c.Put(1, 10)
	if c.Len() != 1 {
		t.Error("Expected length 1")
	}

	c.Put(2, 20)
	if c.Len() != 2 {
		t.Error("Expected length 2")
	}

	c.Get(1)
	c.Remove(2)
	if c.Len() != 1 {
		t.Error("Expected length 1")
	}

	c.Put(3, 30)
	if c.Len() != 2 {
		t.Error("Expected length 2")
	}

	c.Put(4, 40)
	if c.Len() != 3 {
		t.Error("Expected length 3")
	}

	c.Get(4)

	c.Put(5, 50)
	if c.Len() != 3 {
		t.Error("Expected length 3")
	}

	if _, err := c.Get(3); err != ErrKeyNotFound {
		t.Error("Expected 3 not to be found")
	}
}

func TestLRUEviction(t *testing.T) {

	e := NewLRUEviction()

	for i := 1; i <= 4; i++ {
		e.Added(i)
	}

	e.Hit(2)
	e.Hit(5)

	if !e.Removed(3) {
		t.Fatalf("should be able to remove 3")
	}
	if e.Removed(6) {
		t.Fatalf("should not be able to remove 6")
	}

	expectedOrder := []interface{}{1, 4, 2, 5}
	for i, exp := range expectedOrder {
		a := e.Pop()
		t.Logf("Pop() => %v", a)
		if a != exp {
			t.Fatalf("Pop() mismatchs (step #%d), expected %v, got %v", i+1, exp, a)
		}
	}
	if e.Pop() != nil {
		t.Fatalf("not empty when it should")
	}
}

func TestLFUEviction(t *testing.T) {

	e := NewLFUEviction()

	for i := 1; i <= 3; i++ {
		e.Added(i)
	}

	e.Hit(2)
	e.Hit(2)
	e.Hit(4)

	if !e.Removed(3) {
		t.Fatalf("should be able to remove 3")
	}
	if e.Removed(5) {
		t.Fatalf("should not be able to remove 5")
	}

	expectedOrder := []interface{}{1, 4, 2}
	for i, exp := range expectedOrder {
		a := e.Pop()
		t.Logf("Pop() => %v", a)
		if a != exp {
			t.Fatalf("Pop() mismatchs (step #%d), expected %v, got %v", i+1, exp, a)
		}
	}
	if e.Pop() != nil {
		t.Fatalf("not empty when it should")
	}
}
