package dto

type SearchFriendReq struct {
	Username string `json:"username"`
}

type SearchFriendResp struct {
	Username string `json:"username"`
	UUID     string `json:"uuid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// 发送好友请求的数据传输格式
type AddFriendReq struct {
	TargetUserName string `json:"target_user_name"`
	Content        string `json:"content"`
}

// 发送好友请求的回复数据传输格式
type AddFriendResp struct {
	OriginUUID     string `json:"uuid"`
	TargetUserName string `json:"target_user_name"`
}

// 好友列表的数据传输格式
type FriendInfo struct {
	FriendUUID     string `json:"friend_uuid"`
	FriendName     string `json:"friend_name"`
	FriendAvatar   string `json:"friend_avatar"`
	FriendNickname string `json:"friend_nickname"`
	Status         int    `json:"status"`
}

// 处理好友请求
type HandleFriendRequest struct {
	Status     int    `json:"status"` // 0:不同意 1：同意
	TargetUUID string `json:"target_uuid"`
	// AddFriendReq AddFriendReq `json:"add_friend_req"`
}
