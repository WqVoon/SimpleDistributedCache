package main

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	// hash 函数
	hash Hash
	// 记录一个真实节点对应几个虚拟节点
	replicas int
	// 哈希环
	keys []int
	// 用于存储虚拟节点与真实节点的映射，键是虚拟节点的哈希，值是真实节点的名称
	hashMap map[int]string
}

func makeMap(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

/*
将 keys 中的每个键对应的 replicas 个虚拟节点加入到哈希环中
*/
func (self *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < self.replicas; i++ {
			hash := int(self.hash([]byte(strconv.Itoa(i) + key)))
			self.keys = append(self.keys, hash)
			self.hashMap[hash] = key
		}
	}
	sort.Ints(self.keys)
}

/*
根据一致性哈希算法利用 key 计算 self.hashMap 中对应的节点名
*/
func (self *Map) Get(key string) string {
	// 表示当前哈希环上没有任何数据，直接返回
	if len(self.keys) == 0 {
		return ""
	}
	// 根据提供的键来计算相应的哈希
	hash := int(self.hash([]byte(key)))
	// 在哈希环中查找第一个大于等于上面哈希值的下标
	idx := sort.Search(len(self.keys), func(i int) bool {
		return self.keys[i] >= hash
	}) % len(self.keys)
	// 返回相应的节点名，取余是为了当 idx == len(self.keys) 时令其为 0
	return self.hashMap[self.keys[idx]]
}
