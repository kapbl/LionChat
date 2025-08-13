package dto

// 登录请求
type LoginRequestDTO struct {
	Email    string `json:"email"`    // 邮箱
	Password string `json:"password"` // 密码
}
type LoginResponseDTO struct {
	BaseResponse
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data LoginData `json:"data"`
}

type RegisterReq struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type UserInfo struct {
	ID       int    `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

type LoginData struct {
	Token    string   `json:"token"`
	UserInfo UserInfo `json:"userinfo"`
}

type RegisterResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
