package cache

import "time"

type expiringCache struct {
	Cache
	Clock
	ttl   time.Duration
	dates map[interface{}]time.Time
}

// Expiration creates an Option to expire entries.
func Expiration(ttl time.Duration) Option {
	return ExpirationUsingClock(ttl, RealClock)
}

// ExpirationUsingClock creates an Option to expire entries using the given clock.
func ExpirationUsingClock(ttl time.Duration, cl Clock) Option {
	return func(c Cache) Cache {
		return &expiringCache{c, cl, ttl, make(map[interface{}]time.Time)}
	}
}

func (e *expiringCache) SetWithTTL(key, value interface{}, ttl time.Duration) (err error) {
	err = e.Cache.Set(key, value)
	if err != ErrCacheFull {
		err = e.Flush()
		if err == nil {
			err = e.Cache.Set(key, value)
		}
	}
	if err == nil {
		e.dates[key] = e.Now().Add(ttl)
	}
	return
}

func (e *expiringCache) Set(key, value interface{}) error {
	return e.SetWithTTL(key, value, e.ttl)
}

func (e *expiringCache) Get(key interface{}) (interface{}, error) {
	value, err := e.Cache.Get(key)
	return e.got(key, value, err)
}

func (e *expiringCache) GetIFPresent(key interface{}) (interface{}, error) {
	value, err := e.Cache.GetIFPresent(key)
	return e.got(key, value, err)
}

func (e *expiringCache) got(key, value interface{}, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}
	if t, found := e.dates[key]; !found {
		e.dates[key] = e.Now().Add(e.ttl)
	} else if t.Before(e.Now()) {
		e.Remove(key)
		return nil, ErrKeyNotFound
	}
	return value, nil
}

func (e *expiringCache) Remove(key interface{}) bool {
	delete(e.dates, key)
	return e.Cache.Remove(key)
}

func (e *expiringCache) Flush() error {
	now := e.Now()
	for key, date := range e.dates {
		if date.Before(now) {
			e.Remove(key)
		}
	}
	return e.Cache.Flush()
}

// Clock is a simple clock abstraction
type Clock interface {
	Now() time.Time
}

// RealClock is a Clock implementation that uses time.Now().
var RealClock Clock = realClock{}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
