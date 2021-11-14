package lru

import "container/list"

type Lru struct {
	maxBytes int64
	nBytes   int64
	ll       *list.List // 双向链表
	cache    map[string]*list.Element
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
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     map[string]*list.Element{},
		OnEvicted: onEvicted,
	}
}

func (lru *Lru) Add(key string, value Value) {
	// 更新缓存
	if e, ok := lru.cache[key]; ok {
		lru.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		lru.nBytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else { // 新增缓存
		e := lru.ll.PushFront(&entry{key, value})
		lru.nBytes += int64(len(key)) + int64(value.Len())
		lru.cache[key] = e
	}
	for lru.maxBytes != 0 && lru.maxBytes < lru.nBytes {
		lru.RemoveOldest()
	}
}

func (lru *Lru) Get(key string) (Value, bool) {
	if e, ok := lru.cache[key]; ok {
		// 如果存在的话, 调整其在双链表的位置
		lru.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (lru *Lru) RemoveOldest() {
	// 取到节点
	e := lru.ll.Back()
	// 如果不为空才用删除
	if e != nil {
		// 从双链表中删除
		lru.ll.Remove(e)
		kv := e.Value.(*entry)
		// 为什么双链表的节点是entry? 就因为方便删除
		delete(lru.cache, kv.key)
		lru.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if lru.OnEvicted != nil {
			lru.OnEvicted(kv.key, kv.value)
		}
	}
}

func (lru *Lru) Len() int {
	return lru.ll.Len()
}
