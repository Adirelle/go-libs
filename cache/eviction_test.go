package cache

import (
	"testing"
)

func TestLRUEviction(t *testing.T) {

	e := newLRUEviction()

	for i := 1; i <= 4; i++ {
		e.Add(i)
	}

	e.Hit(2)
	e.Hit(5)

	if !e.Remove(3) {
		t.Fatalf("should be able to remove 3")
	}
	if e.Remove(6) {
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

	e := newLFUEviction()

	for i := 1; i <= 3; i++ {
		e.Add(i)
	}

	e.Hit(2)
	e.Hit(2)
	e.Hit(4)

	if !e.Remove(3) {
		t.Fatalf("should be able to remove 3")
	}
	if e.Remove(5) {
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
