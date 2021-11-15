package geecache

import (
	"fmt"
	"geeCache/geecache/hashRing"
	"geeCache/geecache/peer"
	"log"
	"net/http"
	"strings"
	"sync"
)

// 注意, 后面还有斜杠
const defaultBasePath = "/_geeCache/"
const defaultReplicas = 50

// implements PeerPicker for a pool of HTTP peers.
type Pool struct {
	// 自己的主机名/IP和端口
	self string
	// 节点间通讯地址的前缀
	basePath string
	sync.Mutex
	Peers       *hashRing.Ring
	HTTPGetters map[string]*peer.HTTPGetter
}

func NewPool(self string) *Pool {
	return &Pool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *Pool) Set(peers ...string) {
	p.Lock()
	defer p.Unlock()
	p.Peers = hashRing.New(defaultReplicas, nil)
	p.Peers.Add(peers...)
	p.HTTPGetters = map[string]*peer.HTTPGetter{}
	for _, pe := range peers {
		p.HTTPGetters[pe] = &peer.HTTPGetter{BaseURL: pe + p.basePath}
	}
}

func (p *Pool) PickPeer(key string) (peer.CacheGetter, bool) {
	p.Lock()
	defer p.Unlock()
	if pe := p.Peers.Get(key); pe != "" && pe != p.self {
		p.Log("Pick pe %s", pe)
		return p.HTTPGetters[pe], true
	}
	return nil, false
}

func (p *Pool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v))
}

func (p *Pool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basePath>/<groupName>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", 400)
		return
	}
	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view)
}
