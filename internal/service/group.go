package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"errors"
	"time"

	"gorm.io/gorm"
)

type GroupService struct {
	UserId   int
	UserUUID string
	Username string
}

func (g *GroupService) CreateGroup(req *dto.CreateGroupReq) error {
	return nil
}

func (g *GroupService) JoinGroup(req *dto.JoinGroupReq) (dto.JoinGroupResp, error) {
	var group model.Group
	err := dao.DB.Table("group").Where("name = ?", req.GroupName).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.JoinGroupResp{}, errors.New("组不存在")
	}
	// 检查用户是否已经加入组
	var userGroup model.GroupMember
	err = dao.DB.Table("group_member").Where("user_id = ? AND group_id = ?", g.UserId, group.Id).First(&userGroup).Error
	if err == nil {
		return dto.JoinGroupResp{}, errors.New("用户已经加入组")
	}
	// 加入组
	err = dao.DB.Table("group_member").Create(&model.GroupMember{
		GroupId:  group.Id,
		UserId:   g.UserId,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
	}).Error
	if err != nil {
		return dto.JoinGroupResp{}, errors.New("加入组失败")
	}
	// 更新组的成员数量
	err = dao.DB.Table("group").Where("id = ?", group.Id).Update("member_count", group.MemberCount+1).Error
	if err != nil {
		return dto.JoinGroupResp{}, errors.New("更新组的成员数量失败")
	}
	return dto.JoinGroupResp{
		GroupID:     group.Id,
		GroupName:   group.Name,
		MemberCount: group.MemberCount,
	}, nil
}

func LeaveGroup() {

}

func (g *GroupService) GetGroupMember(groupId int) ([]string, error) {
	var groupMembers []*model.GroupMember
	err := dao.DB.Table("group_member").Where("group_id = ?", groupId).Find(&groupMembers).Error
	if err != nil {
		return nil, errors.New("获取组的成员失败")
	}
	memberID := []string{}
	// 在查询到的成员中，找到用户UUID
	for _, member := range groupMembers {
		var user model.User
		err := dao.DB.Table("user").Where("id = ?", member.UserId).First(&user).Error
		if err == nil {
			memberID = append(memberID, user.Uuid)
		}
	}
	return memberID, nil
}

// 获取自己加入的群组
func (g *GroupService) GetGroupList() ([]dto.MyGroupsResp, error) {
	currentUserID := g.UserId
	// 联合查询Group表和GroupMember表
	var groups []dto.MyGroupsResp
	err := dao.DB.Table("group").
		Select("group.id as group_id, group.name as group_name").
		Joins("JOIN group_member ON group.id = group_member.group_id").
		Where("group_member.user_id = ? AND group_member.delete_at IS NULL", currentUserID).
		Where("group.delete_at IS NULL").
		Scan(&groups).Error

	if err != nil {
		return nil, err
	}
	myGroups := make([]dto.MyGroupsResp, len(groups))
	for i, v := range groups {
		myGroups[i].GroupID = v.GroupID
		myGroups[i].GroupName = v.GroupName
	}
	return myGroups, nil
}
