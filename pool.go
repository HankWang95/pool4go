package pool4go

import (
	"errors"
	"net"
	"sync"
)

type Pool interface {
	Get() (net.Conn, error)
	Close()
	Len() int
}

type gPool struct {
	mu            sync.RWMutex
	conns         chan net.Conn
	function      function
	maxConner     int
	leisureConner int
}

type function func() (net.Conn, error)

func NewGPool(leisureConn, maxConn int, f function) (Pool, error) {
	var pool *gPool
	pool = &gPool{
		maxConner:     maxConn,
		leisureConner: leisureConn,
		conns:         make(chan net.Conn, maxConn),
		function:      f,
	}
	pool.newConn(leisureConn)
	return pool, nil
}

func (pool *gPool) newConn(n int) error {
	for i := 0; i < n; i++ {
		conn, err := pool.function()
		if err != nil {
			return err
		}
		pool.conns <- conn
	}
	return nil
}

func (pool *gPool) Get() (net.Conn, error) {
	conns := pool.conns
	if conns == nil {
		return nil, errors.New("pool is closed")
	}

	select {
	case conn := <-conns:
		if conn == nil {
			return nil, errors.New("pool is closed")
		}
		return pool.wrapConn(conn), nil

	default:
		n := pool.Len()
		// 如果未达到最大连接数，创建新连接
		if n < pool.maxConner {
			if err := pool.newConn(1); err != nil {
				return nil, err
			}
			return pool.Get()
			// 等待conner空闲后连接
		} else {
			return pool.wrapConn(<-pool.conns), nil
		}
	}
}
func (pool *gPool) putBack(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil. ")
	}

	pool.mu.RLock()
	defer pool.mu.RUnlock()

	if pool.conns == nil {
		return conn.Close()
	}

	select {
	case pool.conns <- conn:
		return nil
	default:
		return conn.Close()
	}

}
func (pool *gPool) Close() {
	pool.mu.Lock()
	conns := pool.conns
	pool.conns = nil
	pool.function = nil
	pool.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
	return
}

func (pool *gPool) Len() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return len(pool.conns)
}
