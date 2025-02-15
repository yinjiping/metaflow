package lru

import (
	"container/list"
)

type Cache64 struct {
	capacity int
	lruList  *list.List
	cache    map[uint64]*list.Element
}

type entry64 struct {
	key   uint64
	value interface{}
}

func NewCache64(maxEntries int) *Cache64 {
	return &Cache64{
		capacity: maxEntries,
		lruList:  list.New(),
		cache:    make(map[uint64]*list.Element),
	}
}

func (c *Cache64) Add(key uint64, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[uint64]*list.Element)
		c.lruList = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(ee)
		ee.Value.(*entry64).value = value
		return
	}
	ele := c.lruList.PushFront(&entry64{key, value})
	c.cache[key] = ele
	if c.lruList.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *Cache64) Get(key uint64) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.lruList.MoveToFront(ele)
		return ele.Value.(*entry64).value, true
	}
	return
}

// Contain will check if a key is in the cache, but not modify the list
func (c *Cache64) Contain(key uint64) bool {
	if c.cache == nil {
		return false
	}
	_, ok := c.cache[key]
	return ok
}

// Peek will return the key value but not modify the list
func (c *Cache64) Peek(key uint64) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		return ele.Value.(*entry64).value, true
	}
	return
}

// Keys returns a slice of all keys, from oldest to newest
func (c *Cache64) Keys() []uint64 {
	keys := make([]uint64, len(c.cache))
	i := 0
	for ele := c.lruList.Back(); ele != nil; ele = ele.Prev() {
		keys[i] = ele.Value.(*entry64).key
		i++
	}
	return keys
}

// Values returns a slice of all values, from oldest to newest
func (c *Cache64) Values() []interface{} {
	values := make([]interface{}, len(c.cache))
	i := 0
	for ele := c.lruList.Back(); ele != nil; ele = ele.Prev() {
		values[i] = ele.Value.(*entry64).value
		i++
	}
	return values
}

func (c *Cache64) Remove(key uint64) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *Cache64) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.lruList.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache64) removeElement(e *list.Element) {
	c.lruList.Remove(e)
	kv := e.Value.(*entry64)
	delete(c.cache, kv.key)
}

func (c *Cache64) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.lruList.Len()
}

func (c *Cache64) Clear() {
	c.lruList = nil
	c.cache = nil
}
