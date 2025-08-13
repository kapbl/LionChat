package dto

type Base struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type BaseRequest struct {
	RequestID string `json:"request_id"`
}

type BaseResponse struct {
	RequestID string `json:"request_id"`
}
