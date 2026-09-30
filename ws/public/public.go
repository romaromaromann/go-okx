package public

import (
	"sync"

	"github.com/romaromaromann/go-okx/ws"
)

type Public struct {
	C *ws.SafeClient
}

var (
	publicClient          *ws.SafeClient
	publicClientSimulated *ws.SafeClient
	publicClientMu        sync.Mutex
)

func NewPublic(simulated bool) *Public {
	publicClientMu.Lock()
	defer publicClientMu.Unlock()

	if simulated {
		if publicClientSimulated == nil {
			publicClientSimulated = ws.NewSafeClient(ws.EndpointPublicSimulated)
		}
		return &Public{C: publicClientSimulated}
	}

	if publicClient == nil {
		publicClient = ws.NewSafeClient(ws.EndpointPublic)
	}
	return &Public{C: publicClient}
}

func (p *Public) Subscribe(args interface{}, handler ws.Handler, handlerError ws.HandlerError) error {
	subscribe := ws.NewOperateSubscribe(args, handler, handlerError)
	return p.C.Operate(subscribe, nil)
}
