package cache

import (
	"fmt"
	"sync"
	"time"
)

type expiringCache struct {
	Cache
	Clock
	ttl   time.Duration
	dates sync.Map
}

// Expiration adds automatic expiration to new entries using the given delay.
func Expiration(ttl time.Duration) Option {
	return ExpirationUsingClock(ttl, RealClock)
}

// ExpirationUsingClock adds automatic expiration to new entries using the given clock.
func ExpirationUsingClock(ttl time.Duration, cl Clock) Option {
	return func(c Cache) Cache {
		return &expiringCache{Cache: c, Clock: cl, ttl: ttl}
	}
}

func (e *expiringCache) PutWithTTL(key, value interface{}, ttl time.Duration) (err error) {
	err = e.Cache.Put(key, value)
	if err == nil {
		e.dates.Store(key, e.Now().Add(ttl))
	}
	return
}

func (e *expiringCache) Put(key, value interface{}) error {
	return e.PutWithTTL(key, value, e.ttl)
}

func (e *expiringCache) Get(key interface{}) (interface{}, error) {
	value, err := e.Cache.Get(key)
	return e.got(key, value, err)
}

func (e *expiringCache) got(key, value interface{}, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}
	t, _ := e.dates.LoadOrStore(key, e.Now().Add(e.ttl))
	if t.(time.Time).Before(e.Now()) {
		e.Remove(key)
		return nil, ErrKeyNotFound
	}
	return value, nil
}

func (e *expiringCache) Remove(key interface{}) bool {
	e.dates.Delete(key)
	return e.Cache.Remove(key)
}

func (e *expiringCache) Flush() error {
	now := e.Now()
	e.dates.Range(func(key, date interface{}) bool {
		if date.(time.Time).Before(now) {
			e.Cache.Remove(key)
		}
		return true
	})
	return e.Cache.Flush()
}

func (e *expiringCache) String() string {
	return fmt.Sprintf("Expiring(%s,%s)", e.Cache, e.ttl)
}

// Clock is a simple clock abstraction to be used with ExpirationUsingClock.
type Clock interface {
	Now() time.Time
}

// RealClock is a Clock implementation that uses time.Now().
var RealClock Clock = realClock{}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
