package pool4go

import (
	"errors"
	"net"
	"sync"
)

var max chan struct{}


type Pool interface {
	Get() (net.Conn, error)
	Close()
	Len() int
}

type gPool struct {
	mu         sync.RWMutex
	conns      chan net.Conn
	function   function
	MaxConner  int
	InitConner int
}

type function func() (net.Conn, error)

func NewGPool(leisureConn, maxConn int, f function) (Pool, error) {
	max = make(chan struct{}, maxConn)

	var pool *gPool
	pool = &gPool{
		MaxConner:  maxConn,
		InitConner: leisureConn,
		conns:      make(chan net.Conn, maxConn),
		function:   f,
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
	// get方法需要获取token，判断是否达到最大连接数
	max <- struct{}{}
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, errors.New("pool is closed")
		}
		return pool.wrapConn(conn), nil

	default:
		conn, err := pool.function()
		if err != nil {
			return nil, err
		}
		return pool.wrapConn(conn), nil

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
