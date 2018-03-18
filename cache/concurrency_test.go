package cache

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

type delaying struct{ Cache }

func (d delaying) Set(k, v interface{}) error {
	time.Sleep(time.Millisecond * time.Duration(v.(int)))
	return d.Cache.Set(k, v)
}

func (d delaying) Get(k interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * time.Duration(k.(int)))
	return d.Cache.Get(k)
}

func delay(c Cache) Cache {
	return delaying{c}
}

func timedPrintf(t *testing.T) func(string, ...interface{}) {
	ref := time.Now()
	return func(tpl string, args ...interface{}) {
		t.Logf("%s: "+tpl, append([]interface{}{time.Now().Sub(ref)}, args...)...)
	}
}

func TestLockingCache(t *testing.T) {

	c := NewMemoryStorage(Spy(timedPrintf(t)), Locking, delay)

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

func slowRandomLoader(key interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * time.Duration(key.(int)))
	return rand.Int(), nil
}

func doDelayed(milli int, f func() (interface{}, error)) func() (interface{}, error) {
	var (
		value interface{}
		err   error
		wg    sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * time.Duration(milli))
		value, err = f()
	}()
	return func() (interface{}, error) {
		wg.Wait()
		return value, err
	}
}

func TestSingleFlight_Gets(t *testing.T) {
	c := NewLoader(slowRandomLoader, Spy(timedPrintf(t)), SingleFlight)

	af := doDelayed(0, func() (interface{}, error) {
		return c.Get(100)
	})

	bf := doDelayed(50, func() (interface{}, error) {
		return c.Get(100)
	})

	av, aerr := af()
	if aerr != nil {
		t.Fatal("expected no error")
	}
	bv, berr := bf()
	if berr != nil {
		t.Fatal("expected no error")
	}

	if av != bv {
		t.Fatal("expected the same values")
	}
}

func TestSingleFlight_GetAndGetIFPresent(t *testing.T) {
	c := NewLoader(slowRandomLoader, Spy(timedPrintf(t)), SingleFlight)

	af := doDelayed(0, func() (interface{}, error) {
		return c.Get(100)
	})

	bf := doDelayed(50, func() (interface{}, error) {
		return c.GetIFPresent(100)
	})

	av, aerr := af()
	if aerr != nil {
		t.Fatal("expected no error")
	}
	bv, berr := bf()
	if berr != nil {
		t.Fatal("expected no error")
	}

	if av != bv {
		t.Fatal("expected the same values")
	}
}

func TestSingleFlight_GetAndSet(t *testing.T) {

	printf := timedPrintf(t)
	c := NewLoader(slowRandomLoader, Spy(printf), SingleFlight)

	af := doDelayed(1, func() (interface{}, error) {
		return c.Get(100)
	})
	bf := doDelayed(50, func() (interface{}, error) {
		return nil, c.Set(100, 50)
	})

	if _, berr := bf(); berr != nil {
		t.Fatal("expected no error")
	}

	if av, aerr := af(); av != 50 || aerr != nil {
		t.Fatal("expected 50, <nil>")
	}
}

func TestSingleFlight_GetAndRemove(t *testing.T) {

	printf := timedPrintf(t)
	c := NewLoader(slowRandomLoader, Spy(printf), SingleFlight)

	af := doDelayed(1, func() (interface{}, error) {
		return c.Get(100)
	})
	bf := doDelayed(50, func() (interface{}, error) {
		return c.Remove(100), nil
	})

	if br, _ := bf(); !(br.(bool)) {
		t.Fatal("expected true")
	}

	if av, aerr := af(); av != nil || aerr != ErrKeyNotFound {
		t.Fatalf("expected <nil>, %v", ErrKeyNotFound)
	}
}

func TestSingleFlight_Flush(t *testing.T) {

	printf := timedPrintf(t)
	c := NewLoader(slowRandomLoader, Spy(printf), SingleFlight)

	var (
		mu    sync.Mutex
		value interface{}
	)
	go func() {
		mu.Lock()
		defer mu.Unlock()
		value, _ = c.Get(100)
	}()

	time.Sleep(50 * time.Millisecond)
	if err := c.Flush(); err != nil {
		t.Fatal("expected <nil>")
	}

	mu.Lock()
	if value == nil {
		t.Fatal("expected non-nil value")
	}
}
