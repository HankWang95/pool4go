# Pool for Go

## 对象

### 结构


- gPool结构
	- mu            sync.RWMutex
	- conns         chan net.Conn
	- function      function
	- maxConner     int
	- leisureConner int

- Conner结构
	- net.Conn
	- p        *gPool
	- mu       sync.RWMutex
	- unusable bool


### 外部调用

**外部调用NewGPool 返回一个 Pool 接口对象**
    - Get() (net.Conn, error)
	- Close()
	- Len() int
```go
// 新建连接池
	p, _ := NewChannelPool(0, 30, factory) 
// 获取连接对象
	conn, _ := p.Get()
// 使用连接对象(满足net.Conn 接口)
	msg := "hello"
	conn.Write([]byte(msg))
// 关闭（放回池）连接对象
	conn.Close()
// 关闭连接池
	p.Close()
```


