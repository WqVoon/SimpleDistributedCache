package main

import (
	"sync"
)

type SafeCache struct {
	mu         sync.Mutex
	lru        *LRUCache
	cacheBytes int
}

func (self *SafeCache) Add(key string, value Chunk) {
	// 加锁并在函数退出时解锁以保证并发安全
	self.mu.Lock()
	defer self.mu.Unlock()
	// 延迟创建 lru 对象
	if self.lru == nil {
		self.lru = makeLRUCache(self.cacheBytes, nil)
	}
	self.lru.Add(key, value)
}

func (self *SafeCache) Get(key string) (value Chunk, done bool) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.lru == nil {
		return
	}

	if v, ok := self.lru.Get(key); ok {
		return v.(Chunk), ok
	}

	return
}
