package dto

type JoinGroupRequest struct {
	TargetGroup string `json:"target_group"` // 目标群组, 可选
}

type JoinGroupResponse struct {
	BaseResponse
	Code      int       `json:"code"`
	Msg       string    `json:"msg"`
	GroupInfo GroupInfo `json:"group_info"`
}

// new
type CreateGroupRequest struct {
	GroupName   string `json:"group_name"`
	GroupType   string `json:"group_type"`
	Description string `json:"description"`
}

// new
type GroupInfo struct {
	GroupUUID   string `json:"group_uuid"`   // 群组id
	GroupName   string `json:"group_name"`   // 群组名称
	MemberCount int    `json:"member_count"` // 群组成员数量
}

// new
type CreateGroupResponse struct {
	BaseResponse
	Code      int       `json:"code"`
	Msg       string    `json:"msg"`
	GroupInfo GroupInfo `json:"group_info"`
}

type GetGroupsResponse struct {
	BaseResponse
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []GroupInfo `json:"data"`
}

type LeaveGroupRequest struct {
	GroupUUID string `json:"group_uuid"` // 群组id
	GroupName string `json:"group_name"` // 群组名称
}

type LeaveGroupResponse struct {
	BaseResponse
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type GroupMember struct {
	Member      []UserInfo `json:"member"`
	MemberCount int        `json:"member_count"`
	Description string     `json:"description"`
}
type GetGroupMembersResponse struct {
	BaseResponse
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data GroupMember `json:"data"`
}
