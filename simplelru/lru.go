package simplelru

import (
	"container/list"
	"errors"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[Key, Value any] func(key Key, value Value)

// LRU implements a non-thread safe fixed size LRU cache
type LRU[Key comparable, Value any] struct {
	size      int
	evictList *list.List
	items     map[Key]*list.Element
	onEvict   EvictCallback[Key, Value]
}

// entry is used to hold a value in the evictList
type entry[Key, Value any] struct {
	key   Key
	value Value
}

// NewLRU constructs an LRU of the given size
func NewLRU[Key comparable, Value any](size int, onEvict EvictCallback[Key, Value]) (*LRU[Key, Value], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	c := &LRU[Key, Value]{
		size:      size,
		evictList: list.New(),
		items:     make(map[Key]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LRU[Key, Value]) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry[Key, Value]).value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU[Key, Value]) Add(key Key, value Value) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry[Key, Value]).value = value
		return false
	}

	// Add new item
	ent := &entry[Key, Value]{key, value}
	entry := c.evictList.PushFront(ent)
	c.items[key] = entry

	evict := c.evictList.Len() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

// Get looks up a key's value from the cache.
func (c *LRU[Key, Value]) Get(key Key) (value Value, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		if ent.Value.(*entry[Key, Value]) == nil {
			var zeroValue Value
			return zeroValue, false
		}
		return ent.Value.(*entry[Key, Value]).value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU[Key, Value]) Contains(key Key) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU[Key, Value]) Peek(key Key) (value Value, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		return ent.Value.(*entry[Key, Value]).value, true
	}
	return
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU[Key, Value]) Remove(key Key) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU[Key, Value]) RemoveOldest() (key Key, value Value, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
		kv := ent.Value.(*entry[Key, Value])
		return kv.key, kv.value, true
	}
	var zeroKey Key
	var zeroValue Value

	return zeroKey, zeroValue, false
}

// GetOldest returns the oldest entry
func (c *LRU[Key, Value]) GetOldest() (key Key, value Value, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		kv := ent.Value.(*entry[Key, Value])
		return kv.key, kv.value, true
	}

	var zeroKey Key
	var zeroValue Value
	return zeroKey, zeroValue, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU[Key, Value]) Keys() []Key {
	keys := make([]Key, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry[Key, Value]).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU[Key, Value]) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *LRU[Key, Value]) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

// removeOldest removes the oldest item from the cache.
func (c *LRU[Key, Value]) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU[Key, Value]) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry[Key, Value])
	delete(c.items, kv.key)
	if c.onEvict != nil {
		c.onEvict(kv.key, kv.value)
	}
}
