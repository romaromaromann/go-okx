package public

import "github.com/romaromaromann/go-okx/rest/api"

func NewGetFundingRate(param *GetFundingParam) (api.IRequest, api.IResponse) {
	return &api.Request{
		Path:   "/api/v5/public/funding-rate",
		Method: api.MethodGet,
		Param:  param,
	}, &GetFundingResponse{}

}

type GetFundingParam struct {
	InstId string `url:"instId"` // ANY for all symbols
}

type GetFundingResponse struct {
	api.Response
	Data []Funding `json:"data"`
}

type Funding struct {
	FundingRate     string `json:"fundingRate"`
	FundingTime     string `json:"fundingTime"`
	InstId          string `json:"instId"`
	NextFundingTime string `json:"nextFundingTime"`
}
