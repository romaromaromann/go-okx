package public

import (
	"encoding/json"

	"github.com/romaromaromann/go-okx/ws"
)

type HandlerOpenInterest func(EventOpenInterest)

type EventOpenInterest struct {
	Arg  ws.Args        `json:"arg"`
	Data []OpenInterest `json:"data"`
}

type OpenInterest struct {
	InstType string `json:"instType"`
	InstId   string `json:"instId"`
	Oi       string `json:"oi"`
	OiCcy    string `json:"oiCcy"`
	Ts       int64  `json:"ts,string"`
}

func SubscribeOpenInterest(instId string, handler HandlerOpenInterest, handlerError ws.HandlerError, simulated bool) error {
	args := &ws.Args{
		Channel: "open-interest",
		InstId:  instId,
	}

	h := func(message []byte) {
		var event EventOpenInterest
		if err := json.Unmarshal(message, &event); err != nil {
			handlerError(err)
			return
		}
		handler(event)
	}

	return NewPublic(simulated).Subscribe(args, h, handlerError)
}

func SubscribeOpenInterestBatch(argsList []*ws.Args, handler HandlerOpenInterest, handlerError ws.HandlerError, simulated bool) error {
	h := func(message []byte) {
		var event EventOpenInterest
		if err := json.Unmarshal(message, &event); err != nil {
			handlerError(err)
			return
		}
		handler(event)
	}

	return NewPublic(simulated).Subscribe(argsList, h, handlerError)
}
