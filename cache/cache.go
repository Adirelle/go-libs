package cache

import (
	"errors"
	"fmt"
	"reflect"
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

	// GetIFPresent fetchs an entry from the cache, without triggering any automatic LoaderFunc.
	// It returns nil and ErrKeyNotFound when the key is not present.
	GetIFPresent(key interface{}) (interface{}, error)

	// Remove removes an entry from the cache.
	// It returns whether the entry was actually found and removed.
	Remove(key interface{}) bool

	// Flush instructs the cache to perform any pending operations.
	Flush() error

	fmt.Stringer
}

// Option alters the cache behavior, adding new features.
type Option func(Cache) Cache

type options []Option

func (o options) applyTo(c Cache) Cache {
	for i := len(o) - 1; i >= 0; i-- {
		c = o[i](c)
	}
	return c
}

// NewVoidStorage returns a cache that does not store nor return any entries, but can be used for side effects of options.
func NewVoidStorage(opts ...Option) Cache {
	return options(opts).applyTo(voidStorage{})
}

type voidStorage struct{}

func (voidStorage) Set(interface{}, interface{}) error            { return nil }
func (voidStorage) Get(interface{}) (interface{}, error)          { return nil, ErrKeyNotFound }
func (voidStorage) GetIFPresent(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }
func (voidStorage) Remove(interface{}) bool                       { return false }
func (voidStorage) Flush() error                                  { return nil }
func (voidStorage) String() string                                { return "Void()" }

// NewMemoryStorage creates an empty memory storage, using a simple go map.
func NewMemoryStorage(opts ...Option) Cache {
	return options(opts).applyTo(new(memoryStorage))
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

func (s *memoryStorage) String() string {
	return fmt.Sprintf("Memory(%p)", *s)
}

// Printf is printf signature
type Printf func(string, ...interface{})

type spy struct {
	Cache
	f Printf
}

// Spy prints all operation using the provided printf function.
func Spy(f Printf) Option {
	return func(c Cache) Cache {
		return &spy{c, f}
	}
}

func (s *spy) Set(key, value interface{}) (err error) {
	err = s.Cache.Set(key, value)
	s.f("%s.Set(%v, %v) -> %v\n", s.Cache, key, value, err)
	return
}

func (s *spy) Get(key interface{}) (value interface{}, err error) {
	value, err = s.Cache.Get(key)
	s.f("%s.Get(%v) -> %v, %v\n", s.Cache, key, value, err)
	return
}

func (s *spy) GetIFPresent(key interface{}) (value interface{}, err error) {
	value, err = s.Cache.GetIFPresent(key)
	s.f("%s.GetIFPresent(%v) -> %v, %v\n", s.Cache, key, value, err)
	return
}

func (s *spy) Remove(key interface{}) (removed bool) {
	removed = s.Cache.Remove(key)
	s.f("%s.Remove(%v) -> %v\n", s.Cache, key, removed)
	return
}

func (s *spy) Flush() (err error) {
	err = s.Cache.Flush()
	s.f("%s.Flush() -> %v\n", s.Cache, err)
	return
}

type writeThrough struct {
	outer Cache
	inner Cache
}

// WriteThrough adds a second-level cache.
func WriteThrough(outer Cache) Option {
	return func(inner Cache) Cache {
		return &writeThrough{outer, inner}
	}
}

func (c *writeThrough) Set(key, value interface{}) (err error) {
	err = c.outer.Set(key, value)
	if err != nil {
		return
	}
	return c.inner.Set(key, value)
}

func (c *writeThrough) Get(key interface{}) (value interface{}, err error) {
	value, err = c.outer.Get(key)
	if err != ErrKeyNotFound {
		return
	}
	value, err = c.inner.Get(key)
	if err == nil {
		err = c.outer.Set(key, value)
	}
	return
}

func (c *writeThrough) GetIFPresent(key interface{}) (value interface{}, err error) {
	value, err = c.outer.GetIFPresent(key)
	if err != ErrKeyNotFound {
		return
	}
	value, err = c.inner.GetIFPresent(key)
	if err == nil {
		err = c.outer.Set(key, value)
	}
	return
}

func (c *writeThrough) Remove(key interface{}) (removed bool) {
	removed = c.outer.Remove(key)
	return c.inner.Remove(key) || removed
}

func (c *writeThrough) Flush() (err error) {
	err = c.outer.Flush()
	if err == nil {
		return c.inner.Flush()
	}
	return
}

func (c *writeThrough) String() string {
	return fmt.Sprintf("WriteThrough(%s,%s)", c.outer, c.inner)
}

// NewLoader creates a cache from a LoaderFunc
func NewLoader(f LoaderFunc, opts ...Option) Cache {
	return options(opts).applyTo(f)
}

// Loader adds a loading mechanism to the cache to generate values on demand.
// Note that: New*(..., Loader(f)) is equivalent to NewLoader(f, WriteThrough(c)).
func Loader(f LoaderFunc) Option {
	return func(c Cache) Cache {
		return &writeThrough{c, f}
	}
}

// LoaderFunc simulates a cache by calling the functions on call to Get.
type LoaderFunc func(interface{}) (interface{}, error)

// Set is a no-op and never fails.
func (LoaderFunc) Set(interface{}, interface{}) error { return nil }

// Get calls the function.
func (l LoaderFunc) Get(key interface{}) (interface{}, error) { return l(key) }

// GetIFPresent is a no-op and always returns ErrKeyNotFound.
func (LoaderFunc) GetIFPresent(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }

// Remove is a no-op and always returns false.
func (LoaderFunc) Remove(interface{}) bool { return false }

// Flush is a no-op and never fails.
func (LoaderFunc) Flush() error { return nil }

func (l LoaderFunc) String() string {
	return fmt.Sprintf("Loader(0x%08x)", reflect.ValueOf(l).Pointer())
}
