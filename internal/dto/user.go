package dto

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
	// ... 其他用户信息字段
}

type LoginData struct {
	Token    string   `json:"token"`
	UserInfo UserInfo `json:"userinfo"`
}

type LoginResp struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data LoginData `json:"data"`
}

type RegisterResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
