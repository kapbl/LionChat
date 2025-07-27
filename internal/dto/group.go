package dto

type JoinGroupReq struct {
	// GroupID   int    `json:"group_id"`   // 群组id, 可选
	GroupName string `json:"group_name"` // 群组名称, 可选
}

type JoinGroupResp struct {
	GroupID     int    `json:"group_id"`     // 群组id
	GroupName   string `json:"group_name"`   // 群组名称
	MemberCount int    `json:"member_count"` // 群组成员数量
}

type CreateGroupReq struct {
	GroupName string `json:"group_name"` // 群组名称, 可选
}

type CreateGroupResp struct {
	GroupID     int    `json:"group_id"`     // 群组id
	GroupName   string `json:"group_name"`   // 群组名称
	MemberCount int    `json:"member_count"` // 群组成员数量
}

type MyGroupsResp struct {
	GroupID   string `json:"group_id"`   // 群组id
	GroupName string `json:"group_name"` // 群组名称
}
