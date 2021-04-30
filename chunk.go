package main

type Chunk struct {
	b []byte
}

func (self Chunk) Len() int {
	// 获取当前 Chunk 中字节切片的长度
	return len(self.b)
}

func (self Chunk) ByteSlice() []byte {
	// 获取 Chunk 中数组的一份拷贝
	c := make([]byte, self.Len())
	copy(c, self.b)
	return c
}

func (self Chunk) String() string {
	// 返回内部字节数组的字符串字面量
	return string(self.b)
}
