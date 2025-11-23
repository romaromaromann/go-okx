package public

import (
    "encoding/json"

    "github.com/romaromaromann/go-okx/ws"
)

type HandlerTickersType func(EventTickersType)

type EventTickersType struct {
    Arg  ws.Args      `json:"arg"`
    Data []TickerType `json:"data"`
}

type TickerType struct {
    InstType    string `json:"instType"`
    InstId      string `json:"instId"`
    Last        string `json:"last"`
    LastSz      string `json:"lastSz"`
    AskPx       string `json:"askPx"`
    AskSz       string `json:"askSz"`
    BidPx       string `json:"bidPx"`
    BidSz       string `json:"bidSz"`
    Open24h     string `json:"open24h"`
    High24h     string `json:"high24h"`
    Low24h      string `json:"low24h"`
    VolCcy24h   string `json:"volCcy24h"`
    Vol24h      string `json:"vol24h"`
    Change24h   string `json:"change24h"`
    ChangePct24h string `json:"changePct24h"`
    SodUtc0     string `json:"sodUtc0"`
    SodUtc8     string `json:"sodUtc8"`
    Ts          int64  `json:"ts,string"`
}

// Подписка на все тикеры по типу инструмента
func SubscribeTickersType(instType string, handler HandlerTickersType, handlerError ws.HandlerError, simulated bool) error {
    args := &ws.Args{
        Channel:  "tickers",
        InstType: instType,
        // Не указываем InstId для подписки на все инструменты типа
    }

    h := func(message []byte) {
        var event EventTickersType
        if err := json.Unmarshal(message, &event); err != nil {
            handlerError(err)
            return
        }
        handler(event)
    }

    return NewPublic(simulated).Subscribe(args, h, handlerError)
}
