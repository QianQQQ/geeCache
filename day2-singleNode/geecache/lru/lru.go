package lru

import (
	"container/list"
	"sync"
)

type Lru struct {
	// 这玩意很有意思, 啥时候用指针啥时候不用
	sync.Mutex
	MaxBytes int64
	NBytes   int64
	Ll       *list.List // 双向链表
	Cache    map[string]*list.Element
	// 当缓存被淘汰时的回调函数
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// 搞这么复杂是因为要更新nBytes
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Lru {
	return &Lru{
		MaxBytes:  maxBytes,
		Ll:        list.New(),
		Cache:     map[string]*list.Element{},
		OnEvicted: onEvicted,
	}
}

func (lru *Lru) Add(key string, value Value) {
	// 更新缓存
	lru.Lock()
	defer lru.Unlock()
	if e, ok := lru.Cache[key]; ok {
		lru.Ll.MoveToFront(e)
		kv := e.Value.(*entry)
		lru.NBytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else { // 新增缓存
		e := lru.Ll.PushFront(&entry{key, value})
		lru.NBytes += int64(len(key)) + int64(value.Len())
		lru.Cache[key] = e
	}
	for lru.MaxBytes != 0 && lru.MaxBytes < lru.NBytes {
		lru.RemoveOldest()
	}
}

func (lru *Lru) Get(key string) (Value, bool) {
	lru.Lock()
	defer lru.Unlock()
	if e, ok := lru.Cache[key]; ok {
		// 如果存在的话, 调整其在双链表的位置
		lru.Ll.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (lru *Lru) RemoveOldest() {
	// 取到节点
	e := lru.Ll.Back()
	// 如果不为空才用删除
	if e != nil {
		// 从双链表中删除
		lru.Ll.Remove(e)
		kv := e.Value.(*entry)
		// 为什么双链表的节点是entry? 就因为方便删除
		delete(lru.Cache, kv.key)
		lru.NBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if lru.OnEvicted != nil {
			lru.OnEvicted(kv.key, kv.value)
		}
	}
}

func (lru *Lru) Len() int {
	return lru.Ll.Len()
}
