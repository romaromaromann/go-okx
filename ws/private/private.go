package private

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/romaromaromann/go-okx/common"
	"github.com/romaromaromann/go-okx/ws"
)

type Private struct {
	Auth common.Auth
	C    *ws.Client
}

var (
	privateClient          *ws.Client
	privateClientSimulated *ws.Client
	privateClientMu        sync.Mutex
)

func NewPrivate(auth common.Auth) *Private {
	privateClientMu.Lock()
	defer privateClientMu.Unlock()

	if auth.Simulated {
		if privateClientSimulated == nil {
			privateClientSimulated = ws.DefaultClientPrivateSimulated
		}
		return &Private{Auth: auth, C: privateClientSimulated}
	}

	if privateClient == nil {
		privateClient = ws.DefaultClientPrivate
	}
	return &Private{Auth: auth, C: privateClient}
}

func (p *Private) Subscribe(args interface{}, handler ws.Handler, handlerError ws.HandlerError) error {
	subscribe := ws.NewOperateSubscribe(args, handler, handlerError)
	return p.C.Operate(subscribe, p.Login)
}

func (p *Private) Login(conn *websocket.Conn) error {
	args := ws.NewArgsLoginFromAuth(p.Auth)
	login := ws.NewOperateLogin(args)
	return p.C.MessageOperate(conn, login)
}
