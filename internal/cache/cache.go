package cache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	val   int
	timer *time.Timer
}

type Store struct {
	memory map[string]*Cache
	mu     sync.RWMutex
}

func (c *Store) Set(key string, val int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ok := validKey(key); ok {
		if cache, exist := c.memory[key]; !exist {
			cache = &Cache{val: val}
			cache.timer = time.AfterFunc(6*time.Second, func() {
				c.Delete(key)
			})
			c.memory[key] = cache

			return nil
		} else {
			cache.val = val

			if cache.timer != nil {
				cache.timer.Reset(6 * time.Second)
			} else {
				cache.timer = time.AfterFunc(6*time.Second, func() {
					c.deleteNoStop(key) // без Stop() і без зайвих перевірок
					fmt.Println("delete:", key)
				})
			}
			return nil

		}
	}

	return fmt.Errorf("Couldnt write %s in memory", key)
}

func (c *Store) Get(key string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ok := validKey(key); ok {
		if cache, exist := c.memory[key]; exist {
			return cache.val, nil
		}

	}

	return 0, fmt.Errorf("Not found in memory")
}

func (c *Store) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !validKey(key) {
		return fmt.Errorf("invalid key")
	}
	if e, ok := c.memory[key]; ok {
		if e.timer != nil {
			e.timer.Stop()
		}
		delete(c.memory, key)
		return nil
	}
	return fmt.Errorf("not found")
}

func (c *Store) deleteNoStop(key string) {
	c.mu.Lock()
	delete(c.memory, key)
	c.mu.Unlock()
}

func validKey(key string) bool {
	return len(key) > 0 && key != " "
}

func New() *Store {
	return &Store{
		memory: make(map[string]*Cache),
		mu:     sync.RWMutex{},
	}
}
