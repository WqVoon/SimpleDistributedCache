package main

import (
	"fmt"
	"sync"
)

/*
该测试程序试图创建 LOOP_TIMES 个协程来进行缓存的更新与读取
	- 如果使用了 LRUCache，那么有可能抛出 map 的并发错误
	- 如果使用了 SafeCache，则不会有这个问题出现
由此验证了 SafeCache 在并发场景下的安全性
*/

const LOOP_TIMES = 4

var cache = SafeCache{cacheBytes: 100}

// var cache = makeLRUCache(100, nil)

var group = sync.WaitGroup{}

func main() {
	group.Add(LOOP_TIMES)

	for i := 0; i < LOOP_TIMES; i++ {
		go work(i)
	}

	group.Wait()
}

func work(ident int) {
	cache.Add("name", Chunk{
		[]byte(fmt.Sprintf("Data%d", ident)),
	})

	if value, ok := cache.Get("name"); ok {
		fmt.Printf("C(%d): %s\n", ident, value)
	}

	group.Done()
}
