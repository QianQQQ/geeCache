package singleFlight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type CallGroup struct {
	sync.Mutex
	m map[string]*call
}

// 动m就要上锁!
func (g *CallGroup) Do(key string, f func() (interface{}, error)) (interface{}, error) {
	g.Lock()
	if g.m == nil {
		g.m = map[string]*call{}
	}
	// 已经有协程在跑了
	if c, ok := g.m[key]; ok {
		g.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := &call{}
	c.wg.Add(1)
	g.m[key] = c
	g.Unlock()

	c.val, c.err = f()
	c.wg.Done()

	g.Lock()
	delete(g.m, key)
	g.Unlock()
	return c.val, c.err
}
