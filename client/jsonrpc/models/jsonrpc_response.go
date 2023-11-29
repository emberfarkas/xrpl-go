package jsonrpcmodels

import (
	"encoding/json"
	"github.com/xyield/xrpl-go/client"
)

type JsonRpcResponse struct {
	Result    json.RawMessage              `json:"result"`
	Warning   string                       `json:"warning,omitempty"`
	Warnings  []client.XRPLResponseWarning `json:"warnings,omitempty"`
	Forwarded bool                         `json:"forwarded,omitempty"`
}

type AnyJson map[string]interface{}

type ApiWarning struct {
	Id      int         `json:"id"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (r JsonRpcResponse) GetResult(v any) error {
	return json.Unmarshal(r.Result, &v)
}
