package cache

import (
	"errors"
	"sync"
)

// ErrKeyNotFound is returned by Cache.Get*() whenever the key is not present in the cache.
var ErrKeyNotFound = errors.New("Key not found")

// ErrCacheFull is returned by Cache.Put() whenever the cache cannot hold more entries.
var ErrCacheFull = errors.New("Cache is full")

// Cache is the main abstraction.
type Cache interface {
	// Set adds an entry to the cache.
	Set(key, value interface{}) error

	// Get fetchs an entry from the cache.
	// It returns nil and ErrKeyNotFound when the key is not present.
	Get(key interface{}) (interface{}, error)

	// GetIFPresent fetchs an entry from the cache, without triggering any automatic loader.
	// It returns nil and ErrKeyNotFound when the key is not present.
	GetIFPresent(key interface{}) (interface{}, error)

	// Remove removes an entry from the cache.
	// It returns whether the entry was actually found and removed.
	Remove(key interface{}) bool

	// Flush instructs the cache to perform any pending operations.
	Flush() error
}

// Option is used to alter to
type Option func(Cache) Cache

// New creates a Cache from the given storage and options.
// The options are applied from last to first (e.g. innermost to outermost).
func New(storage Cache, options ...Option) Cache {
	c := storage
	for i := len(options) - 1; i >= 0; i-- {
		c = options[i](c)
	}
	return c
}

// VoidStorage is a singleton that does not store nor return any entries.
var VoidStorage Cache = voidStorage{}

type voidStorage struct{}

func (voidStorage) Set(interface{}, interface{}) error            { return nil }
func (voidStorage) Get(interface{}) (interface{}, error)          { return nil, ErrKeyNotFound }
func (voidStorage) GetIFPresent(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }
func (voidStorage) Remove(interface{}) bool                       { return false }
func (voidStorage) Flush() error                                  { return nil }

// NewMemoryStorage creates an empty memory storage, using a simple go map.
func NewMemoryStorage() Cache {
	return new(memoryStorage)
}

type memoryStorage map[interface{}]interface{}

func (s *memoryStorage) Set(key interface{}, value interface{}) error {
	if *s == nil {
		*s = make(map[interface{}]interface{})
	}
	(*s)[key] = value
	return nil
}

func (s *memoryStorage) Get(key interface{}) (interface{}, error) {
	if *s != nil {
		if value, found := (*s)[key]; found {
			return value, nil
		}
	}
	return nil, ErrKeyNotFound
}

func (s *memoryStorage) GetIFPresent(key interface{}) (interface{}, error) {
	return s.Get(key)
}

func (s *memoryStorage) Remove(key interface{}) bool {
	if *s != nil {
		if _, found := (*s)[key]; found {
			delete((*s), key)
			return true
		}
	}
	return false
}

func (s *memoryStorage) Flush() error {
	return nil
}

// Loader emulates a cache by dynamically building the values from the given keys on calls to Get.
type Loader func(interface{}) (interface{}, error)

// Set is a no-op and never fails.
func (Loader) Set(interface{}, interface{}) error { return nil }

// Get calls the loader function.
func (l Loader) Get(key interface{}) (interface{}, error) { return l(key) }

// GetIFPresent is a no-op and always returns ErrKeyNotFound.
func (Loader) GetIFPresent(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }

// Remove is a no-op and always returns false.
func (Loader) Remove(interface{}) bool { return false }

// Flush is a no-op and never fails.
func (Loader) Flush() error { return nil }

// lockingCache secures concurrent access to a Cache using a sync.Mutex.
type lockingCache struct {
	Cache
	mu sync.Mutex
}

// Locking adds locking to an existing cache so it becomes safe to use from several goroutines.
var Locking Option = func(c Cache) Cache {
	return &lockingCache{Cache: c}
}

func (c *lockingCache) Set(key interface{}, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Cache.Set(key, value)
}

func (c *lockingCache) Get(key interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Cache.Get(key)
}

func (c *lockingCache) GetIFPresent(key interface{}) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Cache.GetIFPresent(key)
}

func (c *lockingCache) Remove(key interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Cache.Remove(key)
}

func (c *lockingCache) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Cache.Flush()
}
