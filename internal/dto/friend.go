package dto

// 好友列表的数据传输格式
type FriendInfo struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// 处理好友请求
type HandleFriendRequest struct {
	Status         int    `json:"status"` // 0:不同意 1：同意
	TargetUsername string `json:"target_username"`
}

// 处理好友请求回复
type HandleFriendResponse struct {
	BaseResponse
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

// 加好友请求
type AddFriendRequest struct {
	TargetUsername string `json:"target_user_name"`
	Content        string `json:"content"`
}

// 加好友请求回复
type AddFriendResponse struct {
	BaseResponse
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

// 好友列表回复
type FriendListResponse struct {
	BaseResponse
	Code int32        `json:"code"`
	Msg  string       `json:"msg"`
	Data []FriendInfo `json:"data"`
}
