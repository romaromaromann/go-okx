package public

import "github.com/romaromaromann/go-okx/rest/api"

func NewGetInstruments(param *GetInstrumentsParam) (api.IRequest, api.IResponse) {
	return &api.Request{
		Path:   "/api/v5/public/instruments",
		Method: api.MethodGet,
		Param:  param,
	}, &GetInstrumentsResponse{}
}

type GetInstrumentsParam struct {
	InstType string `url:"instType"`
	Uly      string `url:"uly,omitempty"`
	InstId   string `url:"instId,omitempty"`
}

type GetInstrumentsResponse struct {
	api.Response
	Data []Instrument `json:"data"`
}

type Instrument struct {
	InstType     string `json:"instType"`
	InstId       string `json:"instId"`
	BaseCcy      string `json:"baseCcy"`
	QuoteCcy     string `json:"quoteCcy"`
	BaseSz       string `json:"baseSz,omitempty"`
	QuoteSz      string `json:"quoteSz,omitempty"`
	SettleCcy    string `json:"settleCcy,omitempty"`
	CtVal        string `json:"ctVal,omitempty"`
	CtMult       string `json:"ctMult,omitempty"`
	CtValCcy     string `json:"ctValCcy,omitempty"`
	OptType      string `json:"optType,omitempty"`
	Stk          string `json:"stk,omitempty"`
	ListTime     string `json:"listTime"`
	ExpTime      string `json:"expTime,omitempty"`
	TickSz       string `json:"tickSz"`
	LotSz        string `json:"lotSz"`
	MinSz        string `json:"minSz"`
	State        string `json:"state"`
	MaxLmtSz     string `json:"maxLmtSz"`
	MaxMktSz     string `json:"maxMktSz"`
	MaxTwapSz    string `json:"maxTwapSz"`
	MaxIcebergSz string `json:"maxIcebergSz"`
	MaxTriggerSz string `json:"maxTriggerSz"`
	MaxStopSz    string `json:"maxStopSz"`
}
