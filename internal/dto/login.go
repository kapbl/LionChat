package dto

// 登录请求
type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}
type LoginResponse struct {
	// 请求ID, 由服务端生成
	BaseResponse
	Code        int32  `json:"code"`
	AccessToken string `json:"access_token"`
}
