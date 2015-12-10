package app

import (
	"sync"

	"github.com/ironsmile/nedomi/types"
)

type connections struct {
	sync.RWMutex
	conns map[string]types.IncomingConn
}

func newConnections() *connections {
	return &connections{
		conns: make(map[string]types.IncomingConn, 50),
	}
}

func (c *connections) add(conn types.IncomingConn) {
	c.Lock()
	c.conns[conn.RemoteAddr().String()] = conn
	c.Unlock()
}

func (c *connections) Size() int {
	return len(c.conns)
}

func (c *connections) find(key string) (result types.IncomingConn, ok bool) {
	c.RLock()
	result, ok = c.conns[key]
	c.RUnlock()
	return
}

func (c *connections) remove(input types.IncomingConn) {
	c.Lock()
	key := input.RemoteAddr().String()
	// here we check that hte connection is the one that is provided
	// because there is a posibility for a new connection from the same
	// remote address to be created and added, before the previous one is
	// closed(and removed).
	if conn := c.conns[key]; conn == input {
		delete(c.conns, key)
	}
	c.Unlock()
}
