package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/token"
	"errors"
	"time"

	"gorm.io/gorm"
)

type GroupService struct {
	UserId   int
	UserUUID string
	Username string
}

func (g *GroupService) CreateGroup(req *dto.CreateGroupReq) (*dto.CreateGroupResp, error) {
	// 检查该群组是否存在
	// 如果存在，则加入该群组
	// 如果不存在，则创建该群组
	// 将用户的group_version 加一
	var newGroup1 model.Group
	err := dao.DB.Table("group").Where("name=?", req.GroupName).First(&newGroup1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		newGroup := model.Group{
			UUID:        token.GenUUID(req.GroupName),
			Name:        req.GroupName,
			Desc:        req.Description,
			MemberCount: 1,
			Type:        req.GroupType,
			OwnerId:     g.UserId,
			CreateAt:    time.Now(),
			UpdateAt:    time.Now(),
			DeleteAt:    nil,
		}
		err = dao.DB.Table("group").Create(&newGroup).Error
		if err != nil {
			return nil, errors.New("创建组失败")
		}
		user := model.Users{}
		err = dao.DB.Table("users").Where("id = ?", g.UserId).First(&user).Error
		if err != nil {
			return nil, errors.New("用户不存在")
		}
		user.GroupVersion++
		err = dao.DB.Table("users").Where("id = ?", g.UserId).Updates(&user).Error
		if err != nil {
			return nil, errors.New("更新用户的group_version失败")
		}
		// 加入组
		err = dao.DB.Table("group_member").Create(&model.GroupMember{
			GroupId:   newGroup.Id,
			GroupUUID: newGroup.UUID,
			UserId:    g.UserId,
			UserUUID:  g.UserUUID,
			CreateAt:  time.Now(),
			UpdateAt:  time.Now(),
		}).Error
		if err != nil {
			return nil, errors.New("加入组失败")
		}
		return &dto.CreateGroupResp{
			GroupID:   newGroup.Id,
			GroupName: newGroup.Name,
		}, nil
	}
	return &dto.CreateGroupResp{
		GroupID:   newGroup1.Id,
		GroupName: newGroup1.Name,
	}, nil
}

func (g *GroupService) JoinGroup(req *dto.JoinGroupReq) (dto.JoinGroupResp, error) {
	var group model.Group
	err := dao.DB.Table("group").Where("name = ?", req.GroupName).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.JoinGroupResp{}, errors.New("组不存在")
	}
	// 检查用户是否已经加入组
	var userGroup model.GroupMember
	err = dao.DB.Table("group_member").Where("user_id = ? AND group_id = ? AND delete_at IS NULL", g.UserId, group.Id).First(&userGroup).Error
	if err == nil {
		return dto.JoinGroupResp{}, errors.New("用户已经加入组")
	}
	// 加入组
	err = dao.DB.Table("group_member").Create(&model.GroupMember{
		GroupId:   group.Id,
		GroupUUID: group.UUID,
		UserId:    g.UserId,
		UserUUID:  g.UserUUID,
		CreateAt:  time.Now(),
		UpdateAt:  time.Now(),
	}).Error
	if err != nil {
		return dto.JoinGroupResp{}, errors.New("加入组失败")
	}
	// 更新组的成员数量
	err = dao.DB.Table("group").Where("id = ?", group.Id).Update("member_count", group.MemberCount+1).Error
	if err != nil {
		return dto.JoinGroupResp{}, errors.New("更新组的成员数量失败")
	}
	// 更新用户的group_version
	err = dao.DB.Table("users").Where("id = ?", g.UserId).Update("group_version", gorm.Expr("group_version + ?", 1)).Error
	if err != nil {
		return dto.JoinGroupResp{}, errors.New("更新用户的group_version失败")
	}
	return dto.JoinGroupResp{
		GroupID:     group.UUID,
		GroupName:   group.Name,
		MemberCount: group.MemberCount,
	}, nil
}

// LeaveGroup 离开群组，如果是群主则解散群组
func (g *GroupService) LeaveGroup(req *dto.LeaveGroupReq) (dto.LeaveGroupResp, error) {
	// 查找群组
	var group model.Group
	err := dao.DB.Table("group").Where("name = ?", req.GroupName).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.LeaveGroupResp{}, errors.New("群组不存在")
	}
	if err != nil {
		return dto.LeaveGroupResp{}, errors.New("查询群组失败")
	}
	// 检查用户是否在群组中
	var groupMember model.GroupMember
	err = dao.DB.Table("group_member").Where("user_id = ? AND group_id = ? AND delete_at IS NULL", g.UserId, group.Id).First(&groupMember).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.LeaveGroupResp{}, errors.New("您不在该群组中")
	}
	if err != nil {
		return dto.LeaveGroupResp{}, errors.New("查询群成员失败")
	}
	// 检查是否为群主
	if group.OwnerId == g.UserId {
		// 群主离开，解散群组
		return g.dissolveGroup(group.Id)
	} else {
		// 普通成员离开群组
		return g.leaveMemberGroup(group.Id, group.MemberCount)
	}
}

// dissolveGroup 解散群组
func (g *GroupService) dissolveGroup(groupId int) (dto.LeaveGroupResp, error) {
	// 开启事务
	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 删除所有群成员
	err := tx.Table("group_member").Where("group_id = ?", groupId).Update("delete_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return dto.LeaveGroupResp{}, errors.New("删除群成员失败")
	}
	// 删除群组
	err = tx.Table("group").Where("id = ?", groupId).Update("delete_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return dto.LeaveGroupResp{}, errors.New("删除群组失败")
	}
	// 查询操作用户
	user := model.Users{}
	err = tx.Table("users").Where("id = ?", g.UserId).First(&user).Error
	if err != nil {
		tx.Rollback()
		return dto.LeaveGroupResp{}, errors.New("查询用户失败")
	}
	// 更新自己的群组操作状态，用于Redis
	tx.Table("users").Where("id = ?", g.UserId).Update("group_version", gorm.Expr("group_version + ?", 1))
	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return dto.LeaveGroupResp{}, errors.New("提交事务失败")
	}
	return dto.LeaveGroupResp{
		Message: "群组已解散",
	}, nil
}

// leaveMemberGroup 普通成员离开群组
func (g *GroupService) leaveMemberGroup(groupId int, memberCount int) (dto.LeaveGroupResp, error) {
	// 开启事务
	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 删除群成员记录
	err := tx.Table("group_member").Where("user_id = ? AND group_id = ?", g.UserId, groupId).Update("delete_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return dto.LeaveGroupResp{}, errors.New("离开群组失败")
	}
	// 更新群组成员数量
	newMemberCount := memberCount - 1
	if newMemberCount < 0 {
		newMemberCount = 0
	}
	err = tx.Table("group").Where("id = ?", groupId).Update("member_count", newMemberCount).Error
	if err != nil {
		tx.Rollback()
		return dto.LeaveGroupResp{}, errors.New("更新群组成员数量失败")
	}
	// 查询操作用户
	user := model.Users{}
	err = tx.Table("users").Where("id = ?", g.UserId).First(&user).Error
	if err != nil {
		tx.Rollback()
		return dto.LeaveGroupResp{}, errors.New("查询用户失败")
	}
	// 更新自己的群组操作状态，用于Redis
	tx.Table("users").Where("id = ?", g.UserId).Update("group_version", user.GroupVersion+1)
	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return dto.LeaveGroupResp{}, errors.New("提交事务失败")
	}
	return dto.LeaveGroupResp{
		Message: "已成功离开群组",
	}, nil
}

func (g *GroupService) GetGroupMember(groupUUID string) ([]string, error) {
	var groupMembers []*model.GroupMember
	err := dao.DB.Table("group_member").Where("group_uuid = ?", groupUUID).Find(&groupMembers).Error
	if err != nil {
		return nil, errors.New("获取组的成员失败")
	}
	memberID := []string{}
	// 在查询到的成员中，找到用户UUID
	for _, member := range groupMembers {
		var user model.Users
		err := dao.DB.Table("users").Where("id = ?", member.UserId).First(&user).Error
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
		Select("group.uuid as group_uuid, group.name as group_name").
		Joins("JOIN group_member ON group.uuid = group_member.group_uuid").
		Where("group_member.user_id = ? AND group_member.delete_at IS NULL", currentUserID).
		Where("group.delete_at IS NULL").
		Scan(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}
