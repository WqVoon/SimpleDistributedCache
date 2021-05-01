package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var db = map[string]string{
	"Name":  "PsyDuck",
	"Age":   "21",
	"Hobby": "Program",
}

func main() {
	MakeGroup("info", 20, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("Search key", key, "from db")
			time.Sleep(1 * time.Second)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))

	addr := "localhost:8080"
	log.Println("Serve at", addr)
	http.ListenAndServe(addr, makeHTTPPool(addr))
}
