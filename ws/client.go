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

var PingMessage = []byte("ping")

type OperateCallback func(*websocket.Conn) error

type Client struct {
	Endpoint string
	Dialer   *websocket.Dialer
}

type SafeClient struct {
	client *Client
	mu     sync.Mutex
}

// new Client
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

func (c *Client) Operate(operate *Operate, callback OperateCallback) error {
	conn, _, err := c.dial()
	if err != nil {
		return err
	}

	if callback != nil {
		if err := callback(conn); err != nil {
			return err
		}
	}

	if err := c.MessageOperate(conn, operate); err != nil {
		return err
	}

	if operate.Handler != nil {
		ticker := time.NewTicker(PingTimeout)
		go c.keepAlive(conn, ticker)
		go c.messageLoop(conn, operate)
	}

	return nil
}

func (sc *SafeClient) Operate(operate *Operate, callback OperateCallback) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Safe handler
	originalHandler := operate.Handler
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

	return sc.client.Operate(operate, callback)
}

// message operate
func (c *Client) MessageOperate(conn *websocket.Conn, operate *Operate) error {
	if operate.Request == nil {
		return nil
	}
	if err := conn.WriteJSON(operate.Request); err != nil {
		return err
	}
	if err := conn.ReadJSON(&operate.Response); err != nil {
		return err
	}
	return operate.Response.Error()
}

// loop websocket message
func (c *Client) messageLoop(conn *websocket.Conn, operate *Operate) {
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			operate.HandlerError(err)
			return
		}
		operate.Handler(message)
	}
}

// keep websocket alive
func (c *Client) keepAlive(conn *websocket.Conn, ticker *time.Ticker) {
	defer ticker.Stop()
	for {
		<-ticker.C
		deadline := time.Now().Add(PingDeadline)
		if err := conn.WriteControl(websocket.PingMessage, PingMessage, deadline); err != nil {
			return
		}
	}
}

// dial endpoint
func (c *Client) dial() (*websocket.Conn, *http.Response, error) {
	if c.Dialer == nil {
		c.Dialer = websocket.DefaultDialer
	}
	return c.Dialer.Dial(c.Endpoint, nil)
}
