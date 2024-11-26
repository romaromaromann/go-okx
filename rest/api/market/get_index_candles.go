package market

import "github.com/romaromaromann/go-okx/rest/api"

func NewGetIndexCandles(param *GetCandlesParam) (api.IRequest, api.IResponse) {
	return &api.Request{
		Path:   "/api/v5/market/index-candles",
		Method: api.MethodGet,
		Param:  param,
	}, &GetCandlesResponse{}
}
