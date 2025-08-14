package dto

type WebSocketResponse struct {
	BaseResponse
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}
