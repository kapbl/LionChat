package dto

type BaseResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id"`
}
