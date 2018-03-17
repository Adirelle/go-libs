package cache

import "time"

// expiringCache adds an expiration policy on top of an existing cache.
// Expiration times are only kept in memory.
type expiringCache struct {
	// The cache backend
	Cache

	// The time-to-live of added entries
	TTL time.Duration

	// The clock implementation to use.
	Clock

	// dates holds the expiration dates per key.
	dates map[interface{}]time.Time
}

// Expiration creates an Option to expire entries.
func Expiration(ttl time.Duration) Option {
	return ExpirationUsingClock(ttl, RealClock)
}

// ExpirationUsingClock creates an Option to expire entries using the given clock.
func ExpirationUsingClock(ttl time.Duration, cl Clock) Option {
	return func(c Cache) Cache {
		return &expiringCache{c, ttl, cl, make(map[interface{}]time.Time)}
	}
}

// SetWithTTL tries and puts the entry into its backend, sets to expire after the given duration.
// If the backend reports it is full, the expiringCache flushs itself to remove expired entries and tries again.
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

// Set is a synonym to c.SetWithTTL(key, value, c.TTL)
func (e *expiringCache) Set(key, value interface{}) error {
	return e.SetWithTTL(key, value, e.TTL)
}

// Get gets the entry from its backend.
// If the entry is expired, it is removed from the backend and Get returns ErrKeyNotFound.
func (e *expiringCache) Get(key interface{}) (interface{}, error) {
	value, err := e.Cache.Get(key)
	return e.got(key, value, err)
}

// GetIFPresent gets the entry from its backend.
// If the entry is expired, it is removed from the backend and GetIFPresent returns ErrKeyNotFound.
func (e *expiringCache) GetIFPresent(key interface{}) (interface{}, error) {
	value, err := e.Cache.GetIFPresent(key)
	return e.got(key, value, err)
}

func (e *expiringCache) got(key, value interface{}, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}
	if t, found := e.dates[key]; !found {
		e.dates[key] = e.Now().Add(e.TTL)
	} else if t.Before(e.Now()) {
		e.Remove(key)
		return nil, ErrKeyNotFound
	}
	return value, nil
}

// Remove removes the entry from its backend.
func (e *expiringCache) Remove(key interface{}) bool {
	delete(e.dates, key)
	return e.Cache.Remove(key)
}

// Flush removes all expired entries from its backend then flushs it.
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

// FakeClock is a Clock implementation that must be advanced "manually".
type FakeClock time.Time

// NewFakeClock returns an FakeClock instance so Now() returns a non-zero time.Time
func NewFakeClock() *FakeClock {
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
