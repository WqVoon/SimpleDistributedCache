package main

import (
	"fmt"
	"log"
	"sync"
)

//TODO：什么作用？
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (self GetterFunc) Get(key string) ([]byte, error) {
	return self(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache SafeCache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func MakeGroup(name string, cacheBytes int, getter Getter) *Group {
	if getter == nil {
		panic("Nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: SafeCache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (self *Group) Get(key string) (Chunk, error) {
	// 这里处理空值，是为了防止缓存穿透
	if key == "" {
		return Chunk{}, fmt.Errorf("Key is required")
	}
	// 尝试获取一个缓存的 Chunk，如果获取到便返回之
	if v, ok := self.mainCache.Get(key); ok {
		log.Println("Cache Hited")
		return v, nil
	}
	// 否则加载数据到缓存中
	return self.load(key)
}

func (self *Group) load(key string) (value Chunk, err error) {
	// 当前仅从本地获取数据
	return self.getLocally(key)
}

func (self *Group) getLocally(key string) (value Chunk, err error) {
	// 调用 Group 的 Getter 来获取数据
	bytes, err := self.getter.Get(key)
	if err != nil {
		return
	}
	// 将数据保存在 Chunk 中，并将其推入缓存
	value = Chunk{b: bytes}
	self.populateCache(key, value)
	return value, nil
}

func (self *Group) populateCache(key string, value Chunk) {
	self.mainCache.Add(key, value)
}
