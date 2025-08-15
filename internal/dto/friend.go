package dto

// 好友列表的数据传输格式
type FriendInfo struct {
	FriendUUID     string `json:"friend_uuid"`
	FriendName     string `json:"friend_name"`
	FriendAvatar   string `json:"friend_avatar"`
	FriendNickname string `json:"friend_nickname"`
	Status         int    `json:"status"`
	Version        int    `json:"version"`
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
