package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	EndpointPublic            = "wss://ws.okx.com:8443/ws/v5/public"
	EndpointPrivate           = "wss://ws.okx.com:8443/ws/v5/private"
	EndpointBusiness          = "wss://ws.okx.com:8443/ws/v5/business"
	EndpointPublicSimulated   = "wss://wspap.okx.com:8443/ws/v5/public?brokerId=9999"
	EndpointPrivateSimulated  = "wss://wspap.okx.com:8443/ws/v5/private?brokerId=9999"
	EndpointBusinessSimulated = "wss://wspap.okx.com:8443/ws/v5/business?brokerId=9999"

	PingTimeout       = 20 * time.Second
	PingDeadline      = 10 * time.Second
	ReconnectSubDelay = 20 * time.Millisecond
)

var (
	DefaultClientPublic            = NewClient(EndpointPublic)
	DefaultClientPrivate           = NewClient(EndpointPrivate)
	DefaultClientBusiness          = NewClient(EndpointBusiness)
	DefaultClientPublicSimulated   = NewClient(EndpointPublicSimulated)
	DefaultClientPrivateSimulated  = NewClient(EndpointPrivateSimulated)
	DefaultClientBusinessSimulated = NewClient(EndpointBusinessSimulated)

	PingMessage = []byte("ping")
)

type OperateCallback func(*websocket.Conn) error

type subEntry struct {
	id      int64
	handler Handler
}

type Client struct {
	Endpoint string
	Dialer   *websocket.Dialer

	mu       sync.Mutex
	conn     *websocket.Conn
	handlers []*subEntry
	subs     []*Request
	nextId   int64
	started  bool

	writeMu sync.Mutex
}

type SafeClient struct {
	client *Client
	mu     sync.Mutex
}

func NewClient(endpoint string) *Client {
	return &Client{
		Endpoint: endpoint,
	}
}

func NewSafeClient(endpoint string) *SafeClient {
	return &SafeClient{
		client: &Client{
			Endpoint: endpoint,
			Dialer: &websocket.Dialer{
				HandshakeTimeout: 20 * time.Second,
			},
		},
	}
}

func sameRequest(a, b *Request) bool {
	if a == nil || b == nil || a.Op != b.Op {
		return false
	}
	ja, err1 := json.Marshal(a.Args)
	jb, err2 := json.Marshal(b.Args)
	if err1 != nil || err2 != nil {
		return false
	}
	return bytes.Equal(ja, jb)
}

func hasSub(subs []*Request, req *Request) bool {
	for _, s := range subs {
		if sameRequest(s, req) {
			return true
		}
	}
	return false
}

// Operate — переиспользует существующее соединение, если оно открыто.
func (c *Client) Operate(operate *Operate, callback OperateCallback) error {
	c.mu.Lock()

	if operate.Handler != nil {
		operate.Id = c.nextId
		c.nextId++
		c.handlers = append(c.handlers, &subEntry{id: operate.Id, handler: operate.Handler})
	}

	if operate.Request != nil && operate.Request.Op == OpSubscribe {
		if !hasSub(c.subs, operate.Request) {
			c.subs = append(c.subs, operate.Request)
		}
	}

	if c.conn == nil {
		conn, _, err := c.dial()
		if err != nil {
			c.mu.Unlock()
			return err
		}
		c.conn = conn

		if callback != nil {
			if err := callback(conn); err != nil {
				c.mu.Unlock()
				return err
			}
		}

		if !c.started {
			ticker := time.NewTicker(PingTimeout)
			go c.keepAlive(conn, ticker)
			go c.messageLoop(conn)
			c.started = true
		}
	}

	conn := c.conn
	c.mu.Unlock()

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if operate.Request != nil {
		return conn.WriteJSON(operate.Request)
	}
	return nil
}

func (sc *SafeClient) Operate(operate *Operate, callback OperateCallback) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	originalHandler := operate.Handler
	if originalHandler != nil {
		operate.Handler = func(message []byte) {
			defer func() {
				if r := recover(); r != nil {
					if operate.HandlerError != nil {
						operate.HandlerError(fmt.Errorf("panic: %v", r))
					}
				}
			}()
			originalHandler(message)
		}
	}

	return sc.client.Operate(operate, callback)
}

// Unsubscribe — снимает handler и подписку и отправляет unsubscribe на сервер.
func (c *Client) Unsubscribe(operate *Operate) error {
	if operate == nil {
		return nil
	}

	c.mu.Lock()

	if operate.Handler != nil {
		var kept []*subEntry
		for _, e := range c.handlers {
			if e.id != operate.Id {
				kept = append(kept, e)
			}
		}
		c.handlers = kept
	}

	if operate.Request != nil {
		var kept []*Request
		for _, r := range c.subs {
			if !sameRequest(r, operate.Request) {
				kept = append(kept, r)
			}
		}
		c.subs = kept
	}

	conn := c.conn
	c.mu.Unlock()

	if conn == nil || operate.Request == nil {
		return nil
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return conn.WriteJSON(&Request{Op: OpUnsubscribe, Args: operate.Request.Args})
}

func (sc *SafeClient) Unsubscribe(operate *Operate) error {
	if operate == nil {
		return nil
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.client.Unsubscribe(operate)
}

// messageLoop — читает сообщения и раздаёт их во все handlers.
// При ошибке соединения сбрасывает conn и запускает reconnect.
func (c *Client) messageLoop(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			c.onConnectionLost(conn)
			return
		}

		c.mu.Lock()
		entries := append([]*subEntry(nil), c.handlers...)
		c.mu.Unlock()

		for _, e := range entries {
			e.handler(message)
		}
	}
}

func (c *Client) onConnectionLost(conn *websocket.Conn) {
	c.mu.Lock()
	sameConn := c.conn == conn
	if sameConn {
		c.conn = nil
		c.started = false
	}
	c.mu.Unlock()

	if sameConn {
		go c.reconnect()
	}
}

func (c *Client) reconnect() {
	time.Sleep(5 * time.Second)

	c.mu.Lock()
	conn := c.conn
	needStart := false
	if conn == nil {
		var err error
		conn, _, err = c.dial()
		if err != nil {
			c.mu.Unlock()
			go c.reconnect()
			return
		}
		c.conn = conn
		c.started = true
		needStart = true
	}
	subsCopy := append([]*Request(nil), c.subs...)
	c.mu.Unlock()

	for _, sub := range subsCopy {
		select {
		case <-time.After(ReconnectSubDelay):
		default:
		}
		c.writeMu.Lock()
		err := conn.WriteJSON(sub)
		c.writeMu.Unlock()
		if err != nil {
			break
		}
	}

	if needStart {
		ticker := time.NewTicker(PingTimeout)
		go c.keepAlive(conn, ticker)
		go c.messageLoop(conn)
	}
}

func (c *Client) keepAlive(conn *websocket.Conn, ticker *time.Ticker) {
	defer ticker.Stop()
	for {
		<-ticker.C
		c.mu.Lock()
		alive := c.conn == conn
		c.mu.Unlock()
		if !alive {
			return
		}
		deadline := time.Now().Add(PingDeadline)
		if err := conn.WriteControl(websocket.PingMessage, PingMessage, deadline); err != nil {
			conn.Close()
			c.onConnectionLost(conn)
			return
		}
	}
}

func (c *Client) dial() (*websocket.Conn, *http.Response, error) {
	if c.Dialer == nil {
		c.Dialer = websocket.DefaultDialer
	}
	return c.Dialer.Dial(c.Endpoint, nil)
}

// MessageOperate оставлен для совместимости — используется при login (private).
func (c *Client) MessageOperate(conn *websocket.Conn, operate *Operate) error {
	if operate.Request == nil {
		return nil
	}
	return conn.WriteJSON(operate.Request)
}
