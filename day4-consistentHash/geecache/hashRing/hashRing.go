package hashRing

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Ring struct {
	hash     Hash
	replicas int
	keys     []int // sorted
	hashMap  map[int]string
}

func New(replicas int, f Hash) *Ring {
	r := &Ring{
		hash:     f,
		replicas: replicas,
		keys:     []int{},
		hashMap:  map[int]string{},
	}
	if r.hash == nil {
		r.hash = crc32.ChecksumIEEE
	}
	return r
}

// 加的是节点
func (r *Ring) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < r.replicas; i++ {
			hash := int(r.hash([]byte(strconv.Itoa(i) + key)))
			r.keys = append(r.keys, hash)
			r.hashMap[hash] = key
		}
	}
	sort.Ints(r.keys)
}

// 拿的是缓存
func (r *Ring) Get(key string) string {
	n := len(r.keys)
	if n == 0 {
		return ""
	}
	hash := int(r.hash([]byte(key)))
	index := sort.Search(n, func(i int) bool {
		return r.keys[i] >= hash
	})
	return r.hashMap[r.keys[index%n]]
}
