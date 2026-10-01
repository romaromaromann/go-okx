package business

import (
	"sync"

	"github.com/romaromaromann/go-okx/ws"
)

type Business struct {
	C *ws.Client
}

var (
	businessClient          *ws.Client
	businessClientSimulated *ws.Client
	businessClientMu        sync.Mutex
)

func NewBusiness(simulated bool) *Business {
	businessClientMu.Lock()
	defer businessClientMu.Unlock()

	if simulated {
		if businessClientSimulated == nil {
			businessClientSimulated = ws.DefaultClientBusinessSimulated
		}
		return &Business{C: businessClientSimulated}
	}

	if businessClient == nil {
		businessClient = ws.DefaultClientBusiness
	}
	return &Business{C: businessClient}
}

func (p *Business) Subscribe(args interface{}, handler ws.Handler, handlerError ws.HandlerError) error {
	subscribe := ws.NewOperateSubscribe(args, handler, handlerError)
	return p.C.Operate(subscribe, nil)
}

func (p *Business) SubscribeOperate(operate *ws.Operate) error {
	return p.C.Operate(operate, nil)
}

func (p *Business) Unsubscribe(operate *ws.Operate) error {
	return p.C.Unsubscribe(operate)
}

func UnsubscribeOperate(operate *ws.Operate, simulated bool) error {
	return NewBusiness(simulated).Unsubscribe(operate)
}