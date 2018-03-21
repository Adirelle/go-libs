package cache

import (
	"container/heap"
	"container/list"
	"fmt"
)

// EvictionStrategy is used to select entries to evict when the underlying cache is full.
// Most EvictionStrategy are stateful (they track the cached entries) and must not be used by several cache instances.
type EvictionStrategy interface {
	// Added indicates an entry have been added to the underlying cache.
	Added(key interface{})

	// Removed indicates an entry have been removed from the underlying cache.
	Removed(key interface{}) (removed bool)

	// Hit indicates an entry has been retrieved to from the underlying cache.
	Hit(key interface{})

	// Pop selects an entry to evict. It returns either its key, or nil if there is no entry to evict.
	Pop() (key interface{})
}

type evictingCache struct {
	Cache
	maxLen int
	s      EvictionStrategy
}

// Eviction adds a layer to evict entries when the underlying cache is full.
func Eviction(maxLen int, s EvictionStrategy) Option {
	return func(c Cache) Cache {
		return &evictingCache{c, maxLen, s}
	}
}

// LRUEviction adds entry eviction using the Least-Recently-Used strategy
func LRUEviction(maxLen int) Option {
	return func(c Cache) Cache {
		return &evictingCache{c, maxLen, NewLRUEviction()}
	}
}

// LFUEviction adds entry eviction using the Least-Frequently-Used strategy
func LFUEviction(maxLen int) Option {
	return func(c Cache) Cache {
		return &evictingCache{c, maxLen, NewLFUEviction()}
	}
}

func (c *evictingCache) Put(key, value interface{}) (err error) {
	for c.Cache.Len() >= c.maxLen {
		toEvict := c.s.Pop()
		if toEvict == nil {
			break
		}
		if !c.Cache.Remove(toEvict) {
			break
		}
	}
	err = c.Cache.Put(key, value)
	if err == nil {
		c.s.Added(key)
	}
	return nil
}

func (c *evictingCache) Get(key interface{}) (value interface{}, err error) {
	value, err = c.Cache.Get(key)
	if err == nil {
		c.s.Hit(key)
	}
	return
}

func (c *evictingCache) Remove(key interface{}) bool {
	c.s.Removed(key)
	return c.Cache.Remove(key)
}

func (c *evictingCache) String() string {
	return fmt.Sprintf("Evicting(%s,%d,%v)", c.Cache, c.maxLen, c.s)
}

// Least-Recently Used eviction strategy

type lruEviction struct {
	keys     *list.List
	elements map[interface{}]*list.Element
}

// NewLRUEviction creates a new instance of the Least-Recently-Used strategy.
func NewLRUEviction() EvictionStrategy {
	return &lruEviction{list.New(), make(map[interface{}]*list.Element)}
}

func (e *lruEviction) Added(key interface{}) {
	e.elements[key] = e.keys.PushFront(key)
}

func (e *lruEviction) Removed(key interface{}) (found bool) {
	elem, found := e.elements[key]
	if found {
		e.keys.Remove(elem)
		delete(e.elements, key)
	}
	return
}

func (e *lruEviction) Hit(key interface{}) {
	if elem, found := e.elements[key]; found {
		e.keys.MoveToFront(elem)
	} else {
		e.Added(key)
	}
}

func (e *lruEviction) Pop() (key interface{}) {
	elem := e.keys.Back()
	if elem == nil {
		return
	}
	key = e.keys.Remove(elem)
	delete(e.elements, key)
	return
}

// Least-Frequently Used eviction strategy

type lfuEviction struct {
	heap *countHeap
}

// NewLFUEviction creates a new instance of the Least-Frequently-Used strategy.
func NewLFUEviction() EvictionStrategy {
	e := &lfuEviction{&countHeap{make(map[interface{}]int), nil, nil}}
	heap.Init(e.heap)
	return e
}

func (e *lfuEviction) Added(key interface{}) {
	heap.Push(e.heap, key)
}

func (e *lfuEviction) Removed(key interface{}) (found bool) {
	return e.heap.Removed(key)
}

func (e *lfuEviction) Hit(key interface{}) {
	e.heap.Increase(key)
}

func (e *lfuEviction) Pop() (key interface{}) {
	if e.heap.Len() > 0 {
		key = heap.Pop(e.heap)
	}
	return
}

type countHeap struct {
	index  map[interface{}]int
	keys   []interface{}
	counts []int
}

func (h *countHeap) Len() int {
	return len(h.keys)
}

func (h *countHeap) Less(i, j int) bool {
	return h.counts[i] < h.counts[j]
}

func (h *countHeap) Swap(i, j int) {
	h.counts[i], h.counts[j] = h.counts[j], h.counts[i]
	h.keys[i], h.keys[j] = h.keys[j], h.keys[i]
	h.index[h.keys[i]], h.index[h.keys[j]] = i, j
}

func (h *countHeap) Increase(key interface{}) (found bool) {
	i, found := h.index[key]
	if !found {
		h.Push(key)
		i = h.index[key]
	}
	h.counts[i]++
	heap.Fix(h, i)
	return
}

func (h *countHeap) Removed(key interface{}) (found bool) {
	i, found := h.index[key]
	if found {
		heap.Remove(h, i)
	}
	return
}

func (h *countHeap) Push(key interface{}) {
	n := len(h.keys)
	h.counts = append(h.counts, 0)
	h.keys = append(h.keys, key)
	h.index[key] = n
}

func (h *countHeap) Pop() (key interface{}) {
	n := len(h.keys) - 1
	key = h.keys[n]
	h.counts = h.counts[:n]
	h.keys = h.keys[:n]
	delete(h.index, key)
	return
}
