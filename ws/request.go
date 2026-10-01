package ws

import "reflect"

const (
	OpSubscribe   = "subscribe"
	OpUnsubscribe = "unsubscribe"
	OpLogin       = "login"
)

type Request struct {
	Op   string      `json:"op"`
	Args interface{} `json:"args"`
}

func NewRequestSubscribe(args interface{}) *Request {
	return NewRequest(OpSubscribe, args)
}

func NewRequestLogin(args interface{}) *Request {
	return NewRequest(OpLogin, args)
}

func NewRequest(op string, args interface{}) *Request {
	if args != nil {
		kind := reflect.ValueOf(args).Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			return &Request{Op: op, Args: args}
		}
	}
	return &Request{Op: op, Args: []interface{}{args}}
}
