package cache

import (
	"testing"
)

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
