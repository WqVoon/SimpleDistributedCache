package main

import (
	"fmt"
	"log"
	"testing"
)

var database = map[string]string{
	"Name":  "PsyDuck",
	"Age":   "21",
	"Hobby": "Program",
}

func TestGroup(t *testing.T) {
	queryCount := make(map[string]int)
	g := MakeGroup("score", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("Load", key, "from database")
			if value, ok := database[key]; ok {
				queryCount[key] += 1
				return []byte(value), nil
			} else {
				return nil, fmt.Errorf("%s not exists", key)
			}
		},
	))

	for k, v := range database {
		if value, err := g.Get(k); err != nil || value.String() != v {
			log.Fatal("First test failed")
		}

		if _, err := g.Get(k); err != nil || queryCount[k] > 1 {
			log.Fatal("Second test failed")
		}
	}

	if _, err := g.Get("Unknown"); err == nil {
		log.Fatal("Third test failed")
	}
}
