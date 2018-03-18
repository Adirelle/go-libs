package cache

import (
	"container/heap"
	"container/list"
)

// EvictionStrategy is used to select entries to evict when the underlying cache is full.
// Most EvictionStrategy are stateful (they track the cached entries) and must not be used by several cache instances.
type EvictionStrategy interface {
	// Add indicates an entry have been added to the underlying cache.
	Add(key interface{})

	// Remove indicates an entry have been removed from the underlying cache.
	Remove(key interface{}) (removed bool)

	// Hit indicates an entry has been retrieved to from the underlying cache.
	Hit(key interface{})

	// Pop selects an entry to evict. It returns either its key, or nil if there is no entry to evict.
	Pop() (key interface{})
}

type evictingCache struct {
	Cache
	s EvictionStrategy
}

// Eviction adds a layer to evict entries when the underlying cache is full.
func Eviction(s EvictionStrategy) Option {
	return func(c Cache) Cache {
		return &evictingCache{c, s}
	}
}

// LRUEviction adds entry eviction using the Least-Recently-Used strategy
func LRUEviction(c Cache) Cache {
	return &evictingCache{c, NewLRUEviction()}
}

// LFUEviction adds entry eviction using the Least-Frequently-Used strategy
func LFUEviction(c Cache) Cache {
	return &evictingCache{c, NewLFUEviction()}
}

func (c *evictingCache) Set(key, value interface{}) error {
	for true {
		err := c.Cache.Set(key, value)
		if err == nil {
			break
		}
		if err != ErrCacheFull {
			return err
		}
		toEvict := c.s.Pop()
		if toEvict == nil {
			return err
		}
		c.Remove(toEvict)
	}
	c.s.Add(key)
	return nil
}

func (c *evictingCache) Get(key interface{}) (value interface{}, err error) {
	value, err = c.Cache.Get(key)
	if err == nil {
		c.s.Hit(key)
	}
	return
}

func (c *evictingCache) GetIFPresent(key interface{}) (value interface{}, err error) {
	value, err = c.Cache.GetIFPresent(key)
	if err == nil {
		c.s.Hit(key)
	}
	return
}

func (c *evictingCache) Remove(key interface{}) (removed bool) {
	if removed = c.Cache.Remove(key); removed {
		c.s.Remove(key)
	}
	return
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

func (e *lruEviction) Add(key interface{}) {
	e.elements[key] = e.keys.PushFront(key)
}

func (e *lruEviction) Remove(key interface{}) (found bool) {
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
		e.Add(key)
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

func (e *lfuEviction) Add(key interface{}) {
	heap.Push(e.heap, key)
}

func (e *lfuEviction) Remove(key interface{}) (found bool) {
	return e.heap.Remove(key)
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

func (h *countHeap) Remove(key interface{}) (found bool) {
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
