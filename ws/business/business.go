package business

import (
	"github.com/romaromaromann/go-okx/ws"
)

type Business struct {
	C *ws.Client
}

func NewBusiness(simulated bool) *Business {
	public := &Business{
		C: ws.DefaultClientBusiness,
	}
	if simulated {
		public.C = ws.DefaultClientBusinessSimulated
	}
	return public
}

// subscribe
func (p *Business) Subscribe(args interface{}, handler ws.Handler, handlerError ws.HandlerError) error {
	subscribe := ws.NewOperateSubscribe(args, handler, handlerError)
	return p.C.Operate(subscribe, nil)
}
