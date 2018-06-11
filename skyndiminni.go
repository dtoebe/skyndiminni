package skyndiminni

import (
	"errors"
	"sync"
	"time"
)

// Cahe is the object that controls the whole in-memory cache
type Cache struct {
	*cache
}

type cache struct {
	defaultExpr          time.Duration
	checkExpiredInterval time.Duration
	items                map[string]*Item
	mut                  sync.RWMutex
	wg                   *sync.WaitGroup
}

// item is the value of each key value pair
type Item struct {
	Value        interface{}
	Expiration   int64
	creationTime int64
}

const (
	// NoExpiration sets the the key/value pair to live for the duration of the process
	NoExpiration time.Duration = -1
	// DefaultExpiration is just a simple default parameter for the expiration of the key/value pair
	DefaultExpiration time.Duration = 5 * time.Minute
	// CheckExpired is the default time to check for expired key/value pairs
	CheckExpired time.Duration = 10 * time.Minute
)

// NewCache takes default time as time.Duration for default expiration time
// creates the Cahe intance
// starts a goroutine to periodically check to expired keys
// returns the Cache pointer and an error
func NewCache(defaultExpiration time.Duration) (*Cache, error) {
	c := &cache{
		defaultExpr:          defaultExpiration,
		items:                make(map[string]*Item),
		wg:                   new(sync.WaitGroup),
		checkExpiredInterval: CheckExpired,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			c.initExpiration()
			time.Sleep(c.checkExpiredInterval)
		}
	}()

	return &Cache{c}, nil
}

// Close cleans up the goroutines running
func (c *Cache) Close() {
	c.wg.Wait()
}

// Get gets a non expired value based off provided key
func (c *Cache) Get(key string) (*Item, error) {
	c.mut.RLock()
	item := c.items[key]
	if item == nil {
		c.mut.RUnlock()
		return nil, errors.New("key does not exist")
	}
	if item.Expiration >= time.Now().Unix() {
		delete(c.items, key)
		c.mut.RUnlock()
		return nil, errors.New("key does not exist")
	}
	c.mut.RUnlock()
	return item, nil
}

// Set takes the provided key and checks to make sure it does not exist then creates a new key/value pair with expiration time
func (c *Cache) Set(key string, value interface{}, expirationTime int64) (*Item, error) {
	c.mut.Lock()
	item := c.items[key]
	if item != nil {
		c.mut.Unlock()
		return item, errors.New("key already exists")
	}

	item = &Item{
		Value:        value,
		Expiration:   expirationTime,
		creationTime: time.Now().Unix(),
	}
	c.items[key] = item
	c.mut.Unlock()
	return item, nil
}

// Update takes a key/value pair with expiration time and updates existing key
// if setIfNotExist is true will create new key/value if not exists
// if setIfNotExist is false then will return an error that key already exists
func (c *Cache) Update(key string, value interface{}, expirationTime int64, setIfNotExist bool) (*Item, error) {
	return nil, nil
}

func (c *cache) initExpiration() {
	c.mut.Lock()
	for k, v := range c.items {
		if v.Expiration >= time.Now().Unix() {
			delete(c.items, k)
		}
	}
	c.mut.Unlock()
}
