package cache

import (
	"testing"
)

func TestVoidStorage(t *testing.T) {

	c := NewVoidStorage(Spy(t.Logf))

	if c.Put(5, 6) != nil {
		t.Error("Put: expected <nil>")
	}

	if v, err := c.Get(5); v != nil || err != ErrKeyNotFound {
		t.Errorf("Get: expected <nil>, %v", ErrKeyNotFound)
	}

	if c.Remove(5) {
		t.Error("Remove: expected false")
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}
}

func TestMemoryStorage(t *testing.T) {

	c := NewMemoryStorage(Spy(t.Logf))

	if c.Put(5, 6) != nil {
		t.Error("Put: expected <nil>")
	}

	if v, err := c.Get(5); v != 6 || err != nil {
		t.Error("Get: expected 6, <nil>")
	}

	if !c.Remove(5) {
		t.Error("Remove: expected true")
	}

	if v, err := c.Get(5); v != nil || err != ErrKeyNotFound {
		t.Errorf("Get: expected <nil>, %v", ErrKeyNotFound)
	}

	if c.Remove(5) {
		t.Error("Remove: expected false")
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}
}

func TestLoader(t *testing.T) {

	c := NewLoader(
		func(k interface{}) (interface{}, error) {
			t.Logf("Load %v", k)
			return k, nil
		},
		Spy(t.Logf),
	)

	if v, err := c.Get(5); err != nil || v != 5 {
		t.Error("Get: expected 5, <nil>")
	}

	if err := c.Put(5, 6); err != nil {
		t.Error("Put: expected <nil>")
	}

	if v := c.Remove(5); v {
		t.Error("Remove: expected false")
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}
}

func TestLoaderOption(t *testing.T) {

	ch := make(chan Event, 10)
	c := NewMemoryStorage(
		Emitter(ch),
		Loader(func(k interface{}) (interface{}, error) {
			t.Logf("Load %v", k)
			return k.(int) + 10, nil
		}),
		Spy(t.Logf),
	)

	if v, err := c.Get(5); err != nil || v != 15 {
		t.Error("Get: expected 5, <nil>")
	}

	if v := c.Len(); v != 1 {
		t.Error("Len: expected 1")
	}

	if err := c.Put(5, 6); err != nil {
		t.Error("Put: expected <nil>")
	}

	if v, err := c.Get(5); err != nil || v != 6 {
		t.Error("Get: expected 6, <nil>")
	}

	if v := c.Remove(5); !v {
		t.Error("Remove: expected falstruee")
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}
}

func TestEmiter(t *testing.T) {

	ch := make(chan Event, 1)

	c := NewVoidStorage(Emitter(ch), Spy(t.Logf))

	c.Get(5)
	if e := <-ch; e.Type != GET || e.Key != 5 || e.Value != nil || e.Err != ErrKeyNotFound {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Put(5, 6)
	if e := <-ch; e.Type != PUT || e.Key != 5 || e.Value != 6 || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Remove(5)
	if e := <-ch; e.Type != REMOVE || e.Key != 5 || e.Value != false || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Flush()
	if e := <-ch; e.Type != FLUSH || e.Key != nil || e.Value != nil || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}

	c.Len()
	if e := <-ch; e.Type != LEN || e.Key != nil || e.Value != 0 || e.Err != nil {
		t.Errorf("Event mismatch, got %#v", e)
	}
}
