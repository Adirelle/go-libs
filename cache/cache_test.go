package cache

import (
	"testing"
)

func TestVoidStorage(t *testing.T) {

	c := NewVoidStorage(Spy(t.Logf))

	if c.Set(5, 6) != nil {
		t.Error("Set: expected <nil>")
	}

	if v, err := c.Get(5); v != nil || err != ErrKeyNotFound {
		t.Errorf("Get: expected <nil>, %v", ErrKeyNotFound)
	}

	if v, err := c.GetIFPresent(5); v != nil || err != ErrKeyNotFound {
		t.Errorf("GetIFPresent: expected <nil>, %v", ErrKeyNotFound)
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

	if c.Set(5, 6) != nil {
		t.Error("Set: expected <nil>")
	}

	if v, err := c.Get(5); v != 6 || err != nil {
		t.Error("Get: expected 6, <nil>")
	}

	if v, err := c.GetIFPresent(5); v != 6 || err != nil {
		t.Error("GetIFPresent: expected 6, <nil>")
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

	if err := c.Set(5, 6); err != nil {
		t.Error("Set: expected <nil>")
	}

	if v, err := c.GetIFPresent(5); v != nil || err != ErrKeyNotFound {
		t.Errorf("GetIFPresent: expected <nil>, %s", ErrKeyNotFound)
	}

	if v := c.Remove(5); v {
		t.Error("Remove: expected false")
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}
}
