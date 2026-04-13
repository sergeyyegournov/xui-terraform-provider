package xui

import "encoding/json"

// APIResponse is the standard 3x-ui JSON envelope (see web/entity.Msg).
type APIResponse struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
}
