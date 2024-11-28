package trade

import "github.com/romaromaromann/go-okx/rest/api"

func NewPostAlgoOrder(param *PostAlgoOrderParam) (api.IRequest, api.IResponse) {
	return &api.Request{
		Path:   "/api/v5/trade/order-algo",
		Method: api.MethodPost,
		Param:  param,
	}, &PostAlgoOrderResponse{}
}

type PostAlgoOrderParam struct {
	InstId          string  `json:"instId"`
	TdMode          string  `json:"tdMode"`
	Side            string  `json:"side"`
	PosSide         string  `json:"posSide,omitempty"`
	OrdType         string  `json:"ordType"`
	Sz              string  `json:"sz"`
	SLTriggerPx     float64 `json:"slTriggerPx,omitempty"`
	SlTriggerPxType string  `json:"slTriggerPxType,omitempty,string"`
	SLOrdPx         float64 `json:"slOrdPx,omitempty,string"`
	CxlOnClosePos   bool    `json:"cxlOnClosePos,omitempty"`
	ReduceOnly      bool    `json:"reduceOnly,omitempty"`
	CloseFraction   float64 `json:"closeFraction,omitempty"`
}

type PostAlgoOrderResponse struct {
	api.Response
	Data []AlgoOrder `json:"data"`
}

type AlgoOrder struct {
	OrdId   string `json:"ordId"`
	ClOrdId string `json:"clOrdId"`
	ReqId   string `json:"reqId"`
	SCode   string `json:"sCode"`
	SMsg    string `json:"sMsg"`
}
