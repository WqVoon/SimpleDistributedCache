package main

import "sync"

type Once struct {
	mu sync.Mutex
	m  map[string]*call
}

/* 用于处理 Once 中的一个 string 对应的调用 */
type call struct {
	// 这里采用 WaitGroup 的原因在于并发量很大时会有多次函数调用等待同一个调用结果
	wg sync.WaitGroup
	// 函数调用返回的值
	val Chunk
	err error
}

func (self *Once) Do(key string, fn func() (Chunk, error)) (Chunk, error) {
	self.mu.Lock()
	if self.m == nil {
		self.m = make(map[string]*call)
	}
	// 这里说明以 key 为参数的函数调用曾发生过
	if c, ok := self.m[key]; ok {
		self.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// 否则这里是第一次调用
	c := new(call)
	c.wg.Add(1)
	self.m[key] = c
	self.mu.Unlock()

	// 真正地调用并获取返回值，同时通知上面的 if 中等待的调用们
	c.val, c.err = fn()
	c.wg.Done()

	// 删除掉这次调用的结果
	self.mu.Lock()
	delete(self.m, key)
	self.mu.Unlock()

	return c.val, c.err
}
