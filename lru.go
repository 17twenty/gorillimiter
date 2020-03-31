package gorillimiter

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

// Cache is a concurrent access LRU cache as it locks when mutations are made
type Cache struct {

	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	// how long of a period of time does the rate limit apply
	ratePeriod time.Duration

	evictList *list.List
	// the internal data structure is a map of string -> elements
	// an interface{} would work as well but that's gonna be
	// more expensive at runtime
	cache map[string]*list.Element

	lock sync.RWMutex
}

type entry struct {
	key   string
	value uint64
	// stores the time the entry was first incremented
	updated time.Time
}

// NewLRU gives us a new Least Recently Used Cache
// ratePeriod is the window between now and seconds ago the rate limit applies
func NewLRU(maxEntries int, ratePeriod time.Duration) (*Cache, error) {
	if maxEntries <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	return &Cache{
		MaxEntries: maxEntries,
		evictList:  list.New(),
		cache:      make(map[string]*list.Element),
		ratePeriod: ratePeriod,
	}, nil
}

// Inc allows you to increment a key, if it's over the rate limit maxValue and it's been shorter
// than the grace period then it will return false for the underRateLimit boolean
func (c *Cache) Inc(key string, maxValue int) (uint64, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	underRateLimit := true

	// check to make sure we have space, if not purge the oldest item
	if c.evictList.Len() > c.MaxEntries-1 {
		c.removeOldest()
	}

	if ee, ok := c.cache[key]; ok {
		c.evictList.MoveToFront(ee)
		ee.Value.(*entry).value++
		if ee.Value.(*entry).value > uint64(maxValue) {

			// check to see if we're over our rate limit AND we're within the ratePeriod duration
			// if so then fail the rate limit otherwise reset the times and values for the current period
			if c.ratePeriod > 0 {
				dur := time.Now().UTC().Sub(ee.Value.(*entry).updated)
				if dur > c.ratePeriod {
					ee.Value.(*entry).value = 1
					ee.Value.(*entry).updated = time.Now().UTC()
				} else {
					underRateLimit = false
				}
			} else {
				underRateLimit = false
			}
		}
		return ee.Value.(*entry).value, underRateLimit
	}

	// new item
	item := &entry{key, uint64(1), time.Now().UTC()}

	entry := c.evictList.PushFront(item)
	c.cache[key] = entry
	return item.value, underRateLimit

}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key string) (uint64, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if ent, ok := c.cache[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return 0, false
}

// Remove removes the provided key from the cache
func (c *Cache) Remove(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ent, ok := c.cache[key]; ok {
		c.removeElement(ent)
	}
}

// Len returns the number of items in the cache
func (c *Cache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.evictList.Len()
}

// removeOldest removes the oldest item from cache
func (c *Cache) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *Cache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}
