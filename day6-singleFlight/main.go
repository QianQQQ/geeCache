package main

import (
	"flag"
	"fmt"
	"geeCache/geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("score", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geeCache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := gee.Get(key)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view)
	})
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "geeCache server port")
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], []string(addrs), gee)
}
