package main

import (
	"errors"
	"geeCache/geecache"
	"geeCache/geecache/peer"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, errors.New(key + "not exist")
		}))
	addr := "localhost:8080"
	peers := peer.NewPool(addr)
	log.Println("geeCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
