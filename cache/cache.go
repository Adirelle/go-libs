package cache

import (
	"errors"
	"sync"
)

// ErrKeyNotFound is returned by Cache.Get*() whenever the key is not present in the cache.
var ErrKeyNotFound = errors.New("Key not found")

// Cache is the main abstraction.
type Cache interface {
	// Set adds an entry to the cache.
	Set(key interface{}, value interface{}) error

	// Get fetchs an entry from the cache.
	// It returns nil and ErrKeyNotFound when the key is not present.
	Get(key interface{}) (interface{}, error)

	// GetIFPresent fetchs an entry from the cache, without triggering any automatic loader.
	// It returns nil and ErrKeyNotFound when the key is not present.
	GetIFPresent(key interface{}) (interface{}, error)

	// Remove removes an entry from the cache.
	// It returns whether the entry was actually removed.
	Remove(key interface{}) bool
}

// NullCache does nothing.
type NullCache struct{}

// Set is a no-op.
func (NullCache) Set(interface{}, interface{}) error { return nil }

// Get always returns ErrKeyNotFound.
func (NullCache) Get(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }

// GetIFPresent always returns ErrKeyNotFound.
func (NullCache) GetIFPresent(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }

// Remove always returns false
func (NullCache) Remove(interface{}) bool { return false }

// MemoryStorage is a unbound, lazy-initalized, map-based cache, protected by a lock.
type MemoryStorage map[interface{}]interface{}

// Set stores the key-value pair in the internal map, which is initialized if need be.
func (s *MemoryStorage) Set(key interface{}, value interface{}) error {
	if *s == nil {
		*s = make(map[interface{}]interface{})
	}
	(*s)[key] = value
	return nil
}

// Get fetchs the value from the internal map, returning ErrKeyNotFound if it does not find it.
func (s *MemoryStorage) Get(key interface{}) (interface{}, error) {
	if *s != nil {
		if value, found := (*s)[key]; found {
			return value, nil
		}
	}
	return nil, ErrKeyNotFound
}

// GetIFPresent is a synonym to Get.
func (s *MemoryStorage) GetIFPresent(key interface{}) (interface{}, error) {
	return s.Get(key)
}

// Remove tries and removes the key from the internal map, returning true when it actually removes something.
func (s *MemoryStorage) Remove(key interface{}) bool {
	if *s != nil {
		if _, found := (*s)[key]; found {
			delete((*s), key)
			return true
		}
	}
	return false
}

// Loader emulates a cache to
type Loader func(interface{}) (interface{}, error)

// Set is a no-op.
func (l Loader) Set(interface{}, interface{}) error { return nil }

// Get calls the loader function.
func (l Loader) Get(key interface{}) (interface{}, error) { return l(key) }

// GetIFPresent always returns ErrKeyNotFound.
func (l Loader) GetIFPresent(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }

// Remove always returns false.
func (l Loader) Remove(interface{}) bool { return false }

// LockedCache secures concurrent access to a Cache using a sync.Mutex.
type LockedCache struct {
	Backend Cache
	mu      sync.Mutex
}

// Set acquires the lock before storing the entry into its backend.
func (c *LockedCache) Set(key interface{}, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Backend.Set(key, value)
}

// Get acquires the lock before calling getting the value from its backend.
func (c *LockedCache) Get(key interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Backend.Get(key)
}

// GetIFPresent acquires the lock before calling getting the value from its backend.
func (c *LockedCache) GetIFPresent(key interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Backend.GetIFPresent(key)
}

// Remove acquires the lock before removing the entry from its backend.
func (c *LockedCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Backend.Remove(key)
}
