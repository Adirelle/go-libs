package cache

import (
	"container/heap"
	"container/list"
)

// EvictionStrategy is used by evictingCache to evict items.
// Note that most EvictionStrategy are stateful and must not be shared between several evictingCache.
type EvictionStrategy interface {
	Add(key interface{})
	Remove(key interface{}) (removed bool)
	Hit(key interface{})
	Pop() (key interface{})
}

// evictingCache uses a strategy to evict entries from its backend when the latter is full.
type evictingCache struct {
	Cache
	s EvictionStrategy
}

// LRUEviction adds entry eviction using the Least-Recently-Used strategy
var LRUEviction Option = func(c Cache) Cache {
	return &evictingCache{c, newLRUEviction()}
}

// LFUEviction adds entry eviction using the Least-Frequently-Used strategy
var LFUEviction Option = func(c Cache) Cache {
	return &evictingCache{c, newLFUEviction()}
}

// Set tries to put an entries into its backend. If it is full, it tries to evict an entry.
// It only returns ErrCacheFull if it cannot evict any entries while the backend reports it is full.
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

func newLRUEviction() EvictionStrategy {
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
	key = e.keys.Remove(e.keys.Back())
	delete(e.elements, key)
	return
}

// Least-Frequently Used eviction strategy

type lfuEviction struct {
	heap *countHeap
}

func newLFUEviction() EvictionStrategy {
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
	return heap.Pop(e.heap)
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
	return h.counts[j] < h.counts[i]
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
