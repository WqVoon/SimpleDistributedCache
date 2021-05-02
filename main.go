package main

import (
	"fmt"
	"log"
	"strconv"
)

func main() {
	hash := makeMap(3, func(data []byte) uint32 {
		ret, _ := strconv.Atoi(string(data))
		return uint32(ret)
	})

	// 这时哈希环应该上增加了虚拟节点 2, 4, 12, 14, 22, 24
	hash.Add("4", "2")
	if fmt.Sprintln(hash.keys) != "[2 4 12 14 22 24]\n" {
		log.Fatal("Add func error")
	}

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			log.Fatal("Get func err: ", k, "->", v)
		}
	}

	// 这里加入了 8，哈希环中应该加入了 8, 18, 28，同时 27 应该映射到 8
	hash.Add("8")
	if fmt.Sprintln(hash.keys) != "[2 4 8 12 14 18 22 24 28]\n" {
		log.Fatal("Add func error")
	}
	if hash.Get("27") != "8" {
		log.Fatal("Get func err: 27->8")
	}

	log.Println("All passed")
}
