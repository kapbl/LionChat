package dto

type JoinGroupReq struct {
	// GroupID   int    `json:"group_id"`   // 群组id, 可选
	GroupName string `json:"group_name"` // 群组名称, 可选
}

type JoinGroupResp struct {
	GroupID     string `json:"group_uuid"`   // 群组id
	GroupName   string `json:"group_name"`   // 群组名称
	MemberCount int    `json:"member_count"` // 群组成员数量
}

type CreateGroupReq struct {
	GroupName   string `json:"group_name"` // 群组名称, 可选
	GroupType   string `json:"group_type"`
	Description string `json:"description"`
}

type CreateGroupResp struct {
	GroupID     int    `json:"group_id"`     // 群组id
	GroupName   string `json:"group_name"`   // 群组名称
	MemberCount int    `json:"member_count"` // 群组成员数量
}

type MyGroupsResp struct {
	GroupUUID string `json:"group_uuid"` // 群组id
	GroupName string `json:"group_name"` // 群组名称
}

type LeaveGroupReq struct {
	GroupName string `json:"group_name"` // 群组名称
}

type LeaveGroupResp struct {
	Message string `json:"message"` // 操作结果消息
}
