package main

import "fmt"

// 定义一个 String 类型，让其实现 Value 接口
type String string

func (self String) Len() int {
	return len(self)
}

func main() {
	key1, value1 := "name", String("PsyDuck")
	key2, value2 := "age", String("21")
	key3, value3 := "hobby", String("Program")
	cap := len(key1+key2+key3) + value1.Len() + value2.Len() + value3.Len()
	cache := makeLRUCache(
		cap, func(k string, v Value) {
			fmt.Println(k, "has been removed")
		})

	// 加入了预定的 key 和 value，此时 cache.maxBytes = cache.nBytes
	cache.Add(key1, value1)
	cache.Add(key2, value2)
	cache.Add(key3, value3)
	// 应当能够找到内容，返回 PsyDuck, true
	fmt.Println(cache.Get("name"))
	// 应当能够找到内容，返回 Program, true
	fmt.Println(cache.Get("hobby"))
	// 应当不能找到内容，返回 nil, false
	fmt.Println(cache.Get("location"))
	// 由于占用的内存已经达到最大，此时应当根据 LRU 来删除未被使用过的 age
	cache.Add("*", String("*"))
	// 此时应当无法搜索到 age
	fmt.Println(cache.Get("age"))
	// 但可以搜索到 *
	fmt.Println(cache.Get("*"))
}
