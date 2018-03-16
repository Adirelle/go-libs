package cache

import (
	"container/heap"
	"container/list"
)

// EvictingCache adds an eviction strategy to an existing cache to keep its size under a given strategy.
type EvictingCache struct {
	Backend  Cache
	Strategy EvictionStrategy
}

// EvictionStrategy is used by EvictingCache to evict items.
// Note that EvictionStrategy is stateful and must not be share between several EvictionStrategy.
type EvictionStrategy interface {
	Added(key interface{})
	Removed(key interface{})
	Hit(key interface{})
	ToEvict() (interface{}, bool)
}

func (c *EvictingCache) Set(key interface{}, value interface{}) (err error) {
	err = c.Backend.Set(key, value)
	if err != nil {
		return
	}
	c.Strategy.Added(key)
	key, evict := c.Strategy.ToEvict()
	for evict {
		c.Backend.Remove(key)
		key, evict = c.Strategy.ToEvict()
	}
	return nil
}

func (c *EvictingCache) Get(key interface{}) (value interface{}, err error) {
	value, err = c.Backend.Get(key)
	if err == nil {
		c.Strategy.Hit(key)
	}
	return
}

func (c *EvictingCache) GetIFPresent(key interface{}) (value interface{}, err error) {
	value, err = c.Backend.GetIFPresent(key)
	if err == nil {
		c.Strategy.Hit(key)
	}
	return
}

func (c *EvictingCache) Remove(key interface{}) bool {
	if !c.Backend.Remove(key) {
		return false
	}
	c.Strategy.Removed(key)
	return true
}

// Least-Recently Used eviction strategy

type lruEviction struct {
	capacity int
	keys     *list.List
	elements map[interface{}]*list.Element
}

func NewLRUEviction(capacity int) EvictionStrategy {
	return &lruEviction{capacity, list.New(), make(map[interface{}]*list.Element)}
}

func (e *lruEviction) Added(key interface{}) {
	e.elements[key] = e.keys.PushFront(key)
}

func (e *lruEviction) Removed(key interface{}) {
	if elem, found := e.elements[key]; found {
		e.keys.Remove(elem)
		delete(e.elements, key)
	}
}

func (e *lruEviction) Hit(key interface{}) {
	if elem, found := e.elements[key]; found {
		e.keys.MoveToFront(elem)
	} else {
		e.Added(key)
	}
}

func (e *lruEviction) ToEvict() (key interface{}, evict bool) {
	if evict = e.keys.Len() > e.capacity; evict {
		key = e.keys.Remove(e.keys.Back())
		delete(e.elements, key)
	}
	return
}

// Least-Frequently Used eviction strategy

type lfuEviction struct {
	capacity int
	index    map[interface{}]int
	keys     []interface{}
	counts   []int
}

func NewLFUEviction(capacity int) EvictionStrategy {
	e := &lfuEviction{capacity, make(map[interface{}]int), nil, nil}
	heap.Init(e)
	return e
}

func (e *lfuEviction) Added(key interface{}) {
	heap.Push(e, key)
}

func (e *lfuEviction) Removed(key interface{}) {
	if i, found := e.index[key]; found {
		heap.Remove(e, i)
	}
}

func (e *lfuEviction) Hit(key interface{}) {
	i, found := e.index[key]
	if !found {
		e.Added(key)
		i = e.index[key]
	}
	e.counts[i]++
	heap.Fix(e, i)
}

func (e *lfuEviction) ToEvict() (key interface{}, evict bool) {
	if evict = e.Len() > e.capacity; evict {
		key = heap.Pop(e)
	}
	return
}

func (e *lfuEviction) Len() int {
	return len(e.keys)
}

func (e *lfuEviction) Less(i, j int) bool {
	return e.counts[j] < e.counts[i]
}

func (e *lfuEviction) Swap(i, j int) {
	e.counts[i], e.counts[j] = e.counts[j], e.counts[i]
	e.keys[i], e.keys[j] = e.keys[j], e.keys[i]
	e.index[e.keys[i]], e.index[e.keys[j]] = i, j
}

func (e *lfuEviction) Push(key interface{}) {
	n := len(e.keys)
	e.counts = append(e.counts, 0)
	e.keys = append(e.keys, key)
	e.index[key] = n
}

func (e *lfuEviction) Pop() (key interface{}) {
	n := len(e.keys) - 1
	key = e.keys[n]
	e.counts = e.counts[:n]
	e.keys = e.keys[:n]
	delete(e.index, key)
	return
}
