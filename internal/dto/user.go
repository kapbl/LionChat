package dto

// 获取用户信息响应
type GetUserInfoResponse struct {
	BaseResponse
	Code     int32    `json:"code"`
	UserInfo UserInfo `json:"user_info,omitempty"`
}

// 用户信息
type UserInfo struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// 更新用户信息请求
type UpdateUserReq struct {
	Nickname string `json:"nickname,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	// todo 其他需要更新的字段

}

// 更新用户信息响应
type UpdateUserResponse struct {
	BaseResponse
	Code int32  `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

//搜索用户响应
type SearchClientResponse struct {
	BaseResponse
	Code int32      `json:"code"`
	Data []UserInfo `json:"data,omitempty"`
}
