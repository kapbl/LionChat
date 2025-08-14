package dto

// 注册响应
type RegisterResponse struct {
	// 请求ID, 由服务端生成
	BaseResponse
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

// 注册请求
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}
