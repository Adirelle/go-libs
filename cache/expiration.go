package cache

import "time"

// ExpiringCache adds an expiration policy on top of an existing cache.
type ExpiringCache struct {
	Backend Cache
	TTL     time.Duration
	Clock   Clock
	dates   map[interface{}]time.Time
}

type expiringItem struct {
	value      interface{}
	expiration time.Time
}

func NewExpiringCache(backend Cache, ttl time.Duration) *ExpiringCache {
	return NewExpiringCacheWithClock(backend, ttl, RealClock{})
}

func NewExpiringCacheWithClock(backend Cache, ttl time.Duration, clock Clock) *ExpiringCache {
	return &ExpiringCache{backend, ttl, clock, make(map[interface{}]time.Time)}
}

func (e *ExpiringCache) Set(key interface{}, value interface{}) (err error) {
	err = e.Backend.Set(key, value)
	if err == nil {
		e.dates[key] = e.Clock.Now().Add(e.TTL)
	}
	return
}

func (e *ExpiringCache) Get(key interface{}) (value interface{}, err error) {
	value, err = e.Backend.Get(key)
	if err != nil {
		return
	}
	if t, found := e.dates[key]; !found {
		e.dates[key] = e.Clock.Now().Add(e.TTL)
	} else if t.Before(e.Clock.Now()) {
		e.Remove(key)
		return nil, ErrKeyNotFound
	}
	return
}

func (e *ExpiringCache) GetIFPresent(key interface{}) (value interface{}, err error) {
	value, err = e.Backend.GetIFPresent(key)
	if err != nil {
		return
	}
	if t, found := e.dates[key]; !found {
		e.dates[key] = e.Clock.Now().Add(e.TTL)
	} else if t.Before(e.Clock.Now()) {
		e.Remove(key)
		return nil, ErrKeyNotFound
	}
	return
}

func (e *ExpiringCache) Remove(key interface{}) bool {
	delete(e.dates, key)
	return e.Backend.Remove(key)
}

// Clock is a simple clock abstraction
type Clock interface {
	Now() time.Time
}

// RealClock is a Clock implementation using time.Now().
type RealClock struct{}

// Now returns  time.Now().
func (RealClock) Now() time.Time { return time.Now() }

// FakeClock is a Clock implementation that must be advanced "manually".
// It can be used for testing.
type FakeClock time.Time

// NewFackClock returns an FakeClock instance so Now() returns a non-zero time.Time
func NewFackClock() *FakeClock {
	f := FakeClock(time.Unix(0, 0))
	f.Advance(1)
	return &f
}

// Now returns the current time value.
func (f *FakeClock) Now() time.Time { return time.Time(*f) }

// Advance increase the current time value.
func (f *FakeClock) Advance(d time.Duration) {
	*f = FakeClock(time.Time(*f).Add(d))
}
