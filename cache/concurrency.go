package cache

import (
	"fmt"
	"sync"
)

// locking secures concurrent access to a Cache using a sync.Mutex.
type locking struct {
	Cache
	sync.Mutex
}

// Locking adds locking to an existing cache so it becomes safe to use from several goroutines.
var Locking Option = func(c Cache) Cache {
	return &locking{Cache: c}
}

func (l *locking) Set(key interface{}, value interface{}) error {
	l.Lock()
	defer l.Unlock()
	return l.Cache.Set(key, value)
}

func (l *locking) Get(key interface{}) (interface{}, error) {
	l.Lock()
	defer l.Unlock()
	return l.Cache.Get(key)
}

func (l *locking) GetIFPresent(key interface{}) (interface{}, error) {
	l.Lock()
	defer l.Unlock()
	return l.Cache.GetIFPresent(key)
}

func (l *locking) Remove(key interface{}) bool {
	l.Lock()
	defer l.Unlock()
	return l.Cache.Remove(key)
}

func (l *locking) Flush() error {
	l.Lock()
	defer l.Unlock()
	return l.Cache.Flush()
}

func (l *locking) String() string {
	return fmt.Sprintf("Locked(%s)", l.Cache)
}

type singleFlight struct {
	Cache
	calls map[interface{}]*call
	sync.Mutex
}

// SingleFlight adds a layer to deduplicate queries from concurrent goroutines.
var SingleFlight Option = func(c Cache) Cache {
	return &singleFlight{Cache: c, calls: make(map[interface{}]*call)}
}

func (f *singleFlight) Set(key, value interface{}) (err error) {
	f.Lock()
	defer f.Unlock()
	err = f.Cache.Set(key, value)
	c := f.calls[key]
	if c != nil {
		c.Resolve(value, err)
	}
	return err
}

func (f *singleFlight) Get(key interface{}) (value interface{}, err error) {
	f.Lock()
	c := f.calls[key]
	if c == nil {
		c = newCall(
			func() (interface{}, error) {
				return f.Cache.Get(key)
			},
			func() {
				f.Lock()
				delete(f.calls, key)
				f.Unlock()
			},
		)
		f.calls[key] = c
	}
	f.Unlock()
	return c.Await()
}

func (f *singleFlight) GetIFPresent(key interface{}) (value interface{}, err error) {
	f.Lock()
	c := f.calls[key]
	f.Unlock()
	if c != nil {
		return c.Await()
	}
	return f.Cache.GetIFPresent(key)
}

func (f *singleFlight) Remove(key interface{}) (removed bool) {
	f.Lock()
	c := f.calls[key]
	removed = f.Cache.Remove(key)
	f.Unlock()
	if c != nil {
		c.Resolve(nil, ErrKeyNotFound)
		removed = true
	}
	return removed
}

func (f *singleFlight) Flush() (err error) {
	f.Lock()
	var wg sync.WaitGroup
	wg.Add(len(f.calls))
	for _, c := range f.calls {
		go func(c *call) {
			c.Await()
			wg.Done()
		}(c)
	}
	err = f.Cache.Flush()
	f.Unlock()
	wg.Wait()
	return
}

func (f *singleFlight) String() string {
	return fmt.Sprintf("SingleFlight(%s)", f.Cache)
}

type call struct {
	resolved  bool
	value     interface{}
	err       error
	onResolve func()
	sync.WaitGroup
	sync.Mutex
}

func newCall(process func() (interface{}, error), onResolve func()) *call {
	c := new(call)
	c.onResolve = onResolve
	c.Add(1)
	go func() { c.Resolve(process()) }()
	return c
}

func (c *call) Resolve(value interface{}, err error) {
	c.Lock()
	defer c.Unlock()
	if c.resolved {
		return
	}
	c.resolved = true
	if err != nil {
		c.err = err
	} else {
		c.value = value
	}
	go c.onResolve()
	c.Done()
}

func (c *call) Await() (interface{}, error) {
	c.Wait()
	return c.value, c.err
}
