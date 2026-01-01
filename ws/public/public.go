package public

import (
	"github.com/romaromaromann/go-okx/ws"
)

type Public struct {
	C *ws.SafeClient
}

func NewPublic(simulated bool) *Public {
	endpoint := ws.EndpointPublic
	if simulated {
		endpoint = ws.EndpointPublicSimulated
	}

	return &Public{
		C: ws.NewSafeClient(endpoint), // Новый SafeClient каждый раз
	}
}

// subscribe
func (p *Public) Subscribe(args interface{}, handler ws.Handler, handlerError ws.HandlerError) error {
	subscribe := ws.NewOperateSubscribe(args, handler, handlerError)
	return p.C.Operate(subscribe, nil)
}
