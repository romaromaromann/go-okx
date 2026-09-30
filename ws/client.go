package ws

import (
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

	PingTimeout  = 20 * time.Second
	PingDeadline = 10 * time.Second
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

type Client struct {
	Endpoint string
	Dialer   *websocket.Dialer

	mu       sync.Mutex
	conn     *websocket.Conn
	handlers []Handler
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

// Operate — переиспользует существующее соединение, если оно открыто.
func (c *Client) Operate(operate *Operate, callback OperateCallback) error {
	c.mu.Lock()

	// регистрируем handler сразу — до отправки subscribe,
	// чтобы не потерять первые сообщения
	if operate.Handler != nil {
		c.handlers = append(c.handlers, operate.Handler)
	}

	// соединения нет — открываем
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

	// отправляем subscribe БЕЗ ожидания ответа в этом же conn
	// (ответ придёт в messageLoop и уйдёт во все handlers)
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if operate.Request != nil {
		if err := conn.WriteJSON(operate.Request); err != nil {
			return err
		}
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

// messageLoop — читает ОДИН раз, раздаёт во все handlers.
// При ошибке сбрасывает соединение, чтобы следующий Operate сделал dial.
func (c *Client) messageLoop(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			c.mu.Lock()
			// сбрасываем — следующий Operate переподключится
			if c.conn == conn {
				c.conn = nil
				c.started = false
				c.handlers = nil
			}
			c.mu.Unlock()
			return
		}

		c.mu.Lock()
		handlers := append([]Handler(nil), c.handlers...)
		c.mu.Unlock()

		for _, h := range handlers {
			h(message)
		}
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
