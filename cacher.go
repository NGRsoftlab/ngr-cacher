// Copyright 2020-2024 (c) NGR Softlab

package casher

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

////////////////////////////////////////// Struct and constructor

// Cache struct cache
type Cache struct {
	sync.RWMutex
	items             map[string]Item
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
}

// Item struct cache item
type Item struct {
	Value      interface{}
	Expiration int64
	Created    time.Time
	Options    ItemOptions // experimental
}

// ItemOptions - item type info and on-delete func
type ItemOptions struct {
	setSign      bool
	onDeleteFunc func(item interface{}) error
}

// New Initializing a new memory cache
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	items := make(map[string]Item)

	// cache item
	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

// onDeleteProcessItem - on delete action wrapper (experimental)
func (p *Item) processItemOnDeleteFunc() {
	defer panicRecover()
	if p.Options.setSign {
		err := p.Options.onDeleteFunc(p.Value)
		if err != nil {
			fmt.Println("item onDeleteFunc error: ", err.Error())
		}
	}
}

// panicRecover - recover func for onDelete operations (for example)
func panicRecover() {
	if r := recover(); r != nil {
		fmt.Println("Recovered in function: ", r)
	}
}

////////////////////////////////////////// Get-Set

// Set setting a cache by key
func (c *Cache) Set(key string, value interface{}, duration time.Duration, options ...ItemOptions) {
	var expiration int64

	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()
	defer c.Unlock()

	// set into item options only 1st options object
	itemOptions := ItemOptions{}
	if len(options) > 0 {
		itemOptions.setSign = true
		itemOptions.onDeleteFunc = options[0].onDeleteFunc
	}

	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
		Created:    time.Now(),
		Options:    itemOptions,
	}

}

// Get getting a cache by key
func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.items[key]

	// cache not found
	if !found {
		return nil, false
	}

	if item.Expiration > 0 {
		// cache expired
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}

	}

	return item.Value, true
}

////////////////////////////////////////// Keys-Items

// GetItems returns items list
func (c *Cache) GetItems() (items []string) {
	c.RLock()
	defer c.RUnlock()

	for k := range c.items {
		items = append(items, k)
	}

	return
}

// GetKeys returns key list
func (c *Cache) GetKeys() (keys []string) {
	c.RLock()

	defer c.RUnlock()

	for k := range c.items {
		keys = append(keys, k)
	}

	return
}

// ExpiredKeys returns key list which are expired
func (c *Cache) ExpiredKeys() (keys []string) {
	c.RLock()

	defer c.RUnlock()

	for k, i := range c.items {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}

	return
}

// clearItems removes all the items which key in keys.
func (c *Cache) clearItems(keys []string) {
	c.Lock()

	defer c.Unlock()

	for _, k := range keys {
		item, ok := c.items[k]
		if ok {
			item.processItemOnDeleteFunc()
		}

		delete(c.items, k)
	}
}

////////////////////////////////////////// Cleaning

// StartGC start Garbage Collection
func (c *Cache) StartGC() {
	go c.GC()
}

// GC Garbage Collection
func (c *Cache) GC() {
	for {
		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.ExpiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}

	}
}

// Delete cache by key. Return false if key not found
func (c *Cache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	item, ok := c.items[key]
	if !ok {
		return errors.New("key's not found")
	}

	item.processItemOnDeleteFunc()
	delete(c.items, key)

	return nil
}

// ClearAll - remove all items
func (c *Cache) ClearAll() {
	c.Lock()
	defer c.Unlock()

	for k := range c.items {
		item, ok := c.items[k]

		if ok {
			item.processItemOnDeleteFunc()
		}

		delete(c.items, k)
	}
}
