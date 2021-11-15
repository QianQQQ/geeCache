package peer

import (
	"fmt"
	"geeCache/geecache"
	"log"
	"net/http"
	"strings"
)

// 注意, 后面还有斜杠
const defaultBasePath = "/_geeCache/"

// implements PeerPicker for a pool of HTTP peers.
type Pool struct {
	// 自己的主机名/IP和端口
	self string
	// 节点间通讯地址的前缀
	basePath string
}

func NewPool(self string) *Pool {
	return &Pool{
		self:     self,
		basePath: defaultBasePath,
	}
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
	group := geecache.GetGroup(groupName)
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
