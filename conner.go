package pool4go

import (
	"net"
	"sync"
)

type Conner struct {
	net.Conn
	p        *gPool
	mu       sync.RWMutex
	unusable bool
}

func (pool *gPool) wrapConn(conn net.Conn) net.Conn {
	p := &Conner{p: pool}
	p.Conn = conn
	return p
}

func (conner *Conner) Close() error {
	conner.mu.Lock()
	defer conner.mu.Unlock()

	if conner.unusable {
		if conner.Conn != nil {
			return conner.Conn.Close()
		}
		return nil
	}
	return conner.p.putBack(conner.Conn)
}

func (conner *Conner) MakeUnusable() {
	conner.mu.Lock()
	conner.unusable = true
	conner.mu.Unlock()
}

// todo: 闲置时杀连接调度算法
