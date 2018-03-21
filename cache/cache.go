package cache

import (
	"errors"
	"fmt"
)

// ErrKeyNotFound is returned by Cache.Get*() whenever the key is not present in the cache.
var ErrKeyNotFound = errors.New("Key not found")

// Cache is the main abstraction.
type Cache interface {
	// The string representation should be human-readable. It is used by Spy().
	fmt.Stringer

	// Put stores an entry into the cache.
	Put(key, value interface{}) error

	// Get fetchs an entry from the cache.
	// It returns nil and ErrKeyNotFound when the key is not present.
	Get(key interface{}) (value interface{}, err error)

	// Remove removes an entry from the cache.
	// It returns whether the entry was actually found and removed.
	Remove(key interface{}) bool

	// Flush instructs the cache to finish all pending operations, if any.
	// It must not return before all pending operations are finished.
	Flush() error

	// Len returns the number of entries in the cache.
	Len() int
}

// Option adds optional features new to a cache.
// Please note the order of options is important: they must be listed from outermost to innermost.
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

func (voidStorage) Put(interface{}, interface{}) error   { return nil }
func (voidStorage) Get(interface{}) (interface{}, error) { return nil, ErrKeyNotFound }
func (voidStorage) Remove(interface{}) bool              { return false }
func (voidStorage) Flush() error                         { return nil }
func (voidStorage) Len() int                             { return 0 }
func (voidStorage) String() string                       { return "Void()" }

// NewMemoryStorage creates an empty cache based on a go map.
// It is not safe to use from concurrent goroutines.
func NewMemoryStorage(opts ...Option) Cache {
	return options(opts).applyTo(new(memoryStorage))
}

type memoryStorage map[interface{}]interface{}

func (s *memoryStorage) Put(key interface{}, value interface{}) error {
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

func (s *memoryStorage) Len() int {
	return len(*s)
}

func (s *memoryStorage) String() string {
	return fmt.Sprintf("Memory(%p)", *s)
}

// Printf is a printf-like function to be used with Spy()
type Printf func(string, ...interface{})

type spy struct {
	Cache
	f Printf
}

// Spy logs operations using the given function.
func Spy(f Printf) Option {
	return func(c Cache) Cache {
		return &spy{c, f}
	}
}

func (s *spy) Put(key, value interface{}) (err error) {
	err = s.Cache.Put(key, value)
	s.f("%s.Put(%T(%v), %T(%v)) -> %v", s.Cache, key, key, value, value, err)
	return
}

func (s *spy) Get(key interface{}) (value interface{}, err error) {
	value, err = s.Cache.Get(key)
	s.f("%s.Get(%T(%v)) -> %T(%v), %v", s.Cache, key, key, value, value, err)
	return
}

func (s *spy) Remove(key interface{}) (removed bool) {
	removed = s.Cache.Remove(key)
	s.f("%s.Remove(%T(%v)) -> %v", s.Cache, key, key, removed)
	return
}

func (s *spy) Flush() (err error) {
	err = s.Cache.Flush()
	s.f("%s.Flush() -> %v", s.Cache, err)
	return
}

func (s *spy) Len() (len int) {
	len = s.Cache.Len()
	s.f("%s.Len() -> %v", s.Cache, len)
	return
}

type writeThrough struct {
	outer Cache
	inner Cache
}

// WriteThrough adds a second-level cache.
// Get operations are tried on "outer" first. If it fails, it tries the inner cache.
// If it succeed, the value is written to the outer cache.
// Put and remove operations are forwarded to both caches.
func WriteThrough(outer Cache) Option {
	return func(inner Cache) Cache {
		return &writeThrough{outer, inner}
	}
}

func (c *writeThrough) Put(key, value interface{}) (err error) {
	err = c.outer.Put(key, value)
	if err != nil {
		return
	}
	return c.inner.Put(key, value)
}

func (c *writeThrough) Get(key interface{}) (value interface{}, err error) {
	value, err = c.outer.Get(key)
	if err != ErrKeyNotFound {
		return
	}
	value, err = c.inner.Get(key)
	if err == nil {
		err = c.outer.Put(key, value)
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

func (c *writeThrough) Len() int {
	// Outer only contains a subset of entries of the inner cache.
	return c.inner.Len()
}

func (c *writeThrough) String() string {
	return fmt.Sprintf("WriteThrough(%s,%s)", c.outer, c.inner)
}

// LoaderFunc simulates a cache by calling the functions on call to Get.
type LoaderFunc func(interface{}) (interface{}, error)

type loader struct {
	Cache
	f LoaderFunc
}

// NewLoader creates a pseudo-cache from a LoaderFunc.
func NewLoader(f LoaderFunc, opts ...Option) Cache {
	return options(opts).applyTo(&loader{voidStorage{}, f})
}

// Loader adds a layer to generate values on demand.
func Loader(f LoaderFunc) Option {
	return func(c Cache) Cache {
		return &loader{c, f}
	}
}

func (l *loader) Get(key interface{}) (value interface{}, err error) {
	value, err = l.Cache.Get(key)
	if err != ErrKeyNotFound {
		return
	}
	value, err = l.f(key)
	if err == nil {
		err = l.Cache.Put(key, value)
	}
	return
}

func (l *loader) String() string {
	return fmt.Sprintf("Loader(%s,%v)", l.Cache, l.f)
}

// EventType represents the type of operation that has been performed.
type EventType uint8

// EventType values
const (
	UNKNOWN EventType = iota
	PUT
	GET
	REMOVE
	FLUSH
	LEN
)

func (e EventType) String() string {
	switch e {
	case PUT:
		return "PUT"
	case GET:
		return "GET"
	case REMOVE:
		return "REMOVE"
	case FLUSH:
		return "FLUSH"
	case LEN:
		return "LEN"
	default:
		return fmt.Sprintf("EventType(%d)", e)
	}
}

// GoString is fmt.Sprintf("cache.%s", e)
func (e EventType) GoString() string {
	return fmt.Sprintf("cache.%s", e)
}

// Event represents an operation on a cache.
type Event struct {
	// The type of operation
	Type EventType

	// The targetted cache
	Cache Cache

	// The entry key (PUT, GET, REMOVE)
	Key interface{}

	// The entry value (PUT) or any value returned by the operation (GET, REMOVE, LEN).
	Value interface{}

	// Any error returned by the operation (PUT, GET, FLUSH).
	Err error
}

type emitter struct {
	Cache
	ch chan<- Event
}

// Emitter sends cache events to the given channel.
func Emitter(ch chan<- Event) Option {
	return func(c Cache) Cache {
		return &emitter{c, ch}
	}
}

func (e *emitter) emit(t EventType, key, value interface{}, err error) {
	select {
	case e.ch <- Event{t, e.Cache, key, value, err}:
	default:
	}
}

func (e *emitter) Put(key, value interface{}) (err error) {
	err = e.Cache.Put(key, value)
	e.emit(PUT, key, value, err)
	return
}

func (e *emitter) Get(key interface{}) (value interface{}, err error) {
	value, err = e.Cache.Get(key)
	e.emit(GET, key, value, err)
	return
}

func (e *emitter) Remove(key interface{}) (removed bool) {
	removed = e.Cache.Remove(key)
	e.emit(REMOVE, key, removed, nil)
	return
}

func (e *emitter) Flush() (err error) {
	err = e.Cache.Flush()
	e.emit(FLUSH, nil, nil, err)
	return
}

func (e *emitter) Len() (len int) {
	len = e.Cache.Len()
	e.emit(LEN, nil, len, nil)
	return
}
