package cache

import (
	"sync"
	"testing"
	"time"
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

type delayed struct{ Cache }

func (d delayed) Set(k, v interface{}) error {
	time.Sleep(time.Millisecond * time.Duration(v.(int)))
	return d.Cache.Set(k, v)
}

func (d delayed) Get(k interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * time.Duration(k.(int)))
	return d.Cache.Get(k)
}

func TestLockingCache(t *testing.T) {

	c := NewMemoryStorage(
		Spy(t.Logf),
		Locking,
		func(c Cache) Cache { return delayed{c} },
	)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := c.Set(100, 200); err != nil {
			t.Error("Set: expected <nil>")
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		if v, err := c.Get(100); err != nil || v != 200 {
			t.Error("Get: expected 200, <nil>")
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		if v, err := c.GetIFPresent(100); err != nil || v != 200 {
			t.Error("GetIFPresent: expected 200, <nil>")
		}
	}()

	wg.Wait()

	if !c.Remove(100) {
		t.Error("Remove: expected true")
	}

	if err := c.Flush(); err != nil {
		t.Error("Flush: expected <nil>")
	}

}
