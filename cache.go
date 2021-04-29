package main

import "container/list"

func makeCache(maxBytes int, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nBytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

type Cache struct {
	// 可使用的最大内存量
	maxBytes int
	// 当前已经使用的内存量
	nBytes int
	// 用来存储实际的数据，内部的每一项是一个 Entry 结构体
	ll *list.List
	// map 映射，存储字符串到 Cache.ll 中某一项的映射
	cache map[string]*list.Element
	// 可选函数，在有 Entry 被删除时调用，分别传入 Entry 的两个值
	onEvicted func(string, Value)
}

type Entry struct {
	key   string
	value Value
}

type Value interface {
	// 用来计算值所占的内存空间
	Len() int
}

func (self *Cache) Get(key string) (value Value, done bool) {
	if elm, ok := self.cache[key]; ok {
		// 采用 LRU 算法，因此如果缓存的值被使用过，那么放到队列末尾
		self.ll.MoveToBack(elm)

		// 返回缓存的值
		kv := elm.Value.(*Entry)
		value, done = kv.value, true
	}
	return
}

func (self *Cache) Remove() {
	elm := self.ll.Front()
	if elm != nil {
		kv := elm.Value.(*Entry)

		// 删除队列首部的缓存，队首一定是最长未被访问的
		self.ll.Remove(elm)
		delete(self.cache, kv.key)
		// 减少当前占用的空间
		self.nBytes -= int(len(kv.key)) + int(kv.value.Len())

		if self.onEvicted != nil {
			self.onEvicted(kv.key, kv.value)
		}
	}
}

func (self *Cache) Add(key string, value Value) {
	if elm, ok := self.cache[key]; ok {
		// 如果值已经存在，那么将其移动到队尾（标记为最近使用过）
		// 并更新缓存的内存占用以及对应的值
		self.ll.MoveToBack(elm)
		kv := elm.Value.(*Entry)
		self.nBytes += int(value.Len()) - int(kv.value.Len())
		kv.value = value
	} else {
		// 否则添加一个新的值
		elm := self.ll.PushBack(&Entry{key, value})
		self.cache[key] = elm
		self.nBytes += int(len(key)) + int(value.Len())
	}
	// 删除队首的元素（最长时间未访问的元素），直到所占内存比最大内存小
	for self.maxBytes != 0 && self.maxBytes < self.nBytes {
		self.Remove()
	}
}

func (self *Cache) Len() int {
	// 用来返回当前缓存了多少个键值对
	return self.ll.Len()
}
