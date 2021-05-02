package main

import (
	"fmt"
	"log"
	"sync"
)

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
	peers     PeerPicker
	once      Once
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

/*
初始化除 PeerPicker 外的所有属性，PeerPicker 需要调用 RegisterPeers 来注册
*/
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

/*
注册 Group 的 PeerPicker
*/
func (self *Group) RegisterPeers(peers PeerPicker) {
	if self.peers != nil {
		panic("RegisterPeerPicker called more then once")
	}
	self.peers = peers
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

/*
尝试从自己的 mainCache 中获取 key 对应的缓存；
如果找不到则调用 self.load 利用 Getter 或者从远程获取数据
*/
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

/*
如果自己的 peers 属性为 nil，那么调用 getLocally 从本地获取；
否则调用 peers.PickPeer(key) 获取节点 peer，并调用 getFromPeer(peer, key) 来从远程获取数据
*/
func (self *Group) load(key string) (value Chunk, err error) {
	// 这里虽然 Do 内部的 fn 形成了外包以具备设置最外层的 value 和 err 的能力
	// 但是由于 once 的缘故这里的 fn 不一定会执行，所以获取其返回值并主动赋值给 value 和 err 是必要的
	tmpValue, err := self.once.Do(key, func() (Chunk, error) {
		if self.peers != nil {
			if peer, ok := self.peers.PickPeer(key); ok {
				if value, err = self.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("Failed to get from peer:", err)
			}
		}

		return self.getLocally(key)
	})
	return tmpValue, err
}

/*
调用 peer.Get(self.name, key) 来获取数据
*/
func (self *Group) getFromPeer(peer PeerGetter, key string) (Chunk, error) {
	bytes, err := peer.Get(self.name, key)
	if err != nil {
		return Chunk{}, err
	}
	return Chunk{b: bytes}, nil
}

/*
调用自己的 Getter 来加载数据并保存到缓存中
*/
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
