package dto

type Base struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}
type BaseResponse struct {
	RequestID string `json:"request_id"`
}
