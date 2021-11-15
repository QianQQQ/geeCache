package geecache

import (
	"fmt"
	"geeCache/geecache/lru"
	"geeCache/geecache/peer"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name   string
	getter Getter
	peers  peer.PeerPicker
	*lru.Lru
}

var mu sync.RWMutex
var groups = map[string]*Group{}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name,
		getter,
		nil,
		lru.New(cacheBytes, nil),
	}
	groups[name] = g
	return g
}

func (g *Group) RegisterPeers(peers peer.PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.Lru.Get(key); ok {
		log.Println("[GeeCache] hit")
		return v.(ByteView), nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	if g.peers != nil {
		if p, ok := g.peers.PickPeer(key); ok {
			if value, err := g.getFromPeer(p, key); err == nil {
				return value, nil
			} else {
				log.Println("[geeCache] Failed to get from peer", err)
			}
		}
	}
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return nil, err
	}
	value := ByteView(bytes).ByteSlice()
	g.Lru.Add(key, value)
	return value, nil
}

func (g *Group) getFromPeer(peer peer.CacheGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
