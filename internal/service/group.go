package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/token"
	"errors"
	"time"

	"gorm.io/gorm"
)

type GroupService struct {
	UserID   int
	UserUUID string
	Username string
	DB       *gorm.DB
}

func NewGroupService(userID int, userUUID string, username string, db *gorm.DB) *GroupService {
	return &GroupService{
		UserID:   userID,
		UserUUID: userUUID,
		Username: username,
		DB:       db,
	}
}

// ✅
func (g *GroupService) CreateGroup(req *dto.CreateGroupRequest) (*dto.GroupInfo, *cerror.CodeError) {
	tx := g.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	searchGroup := model.Group{}
	err := tx.Table("group").Where("name=?", req.GroupName).First(&searchGroup).Error
	if searchGroup.Id == 0 {
		newGroup := model.Group{
			UUID:        token.GenUUID(req.GroupName),
			Name:        req.GroupName,
			Desc:        req.Description,
			MemberCount: 1,
			Type:        req.GroupType,
			OwnerId:     g.UserID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			DeletedAt:   nil,
		}
		err = tx.Table("group").Create(&newGroup).Error
		if err != nil {
			return nil, cerror.NewCodeError(1122, "创建组失败")
		}
		// 加入组
		err = tx.Table("group_member").Create(&model.GroupMember{
			GroupId:   newGroup.Id,
			GroupUUID: newGroup.UUID,
			UserId:    g.UserID,
			UserUUID:  g.UserUUID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
		}).Error
		if err != nil {
			return nil, cerror.NewCodeError(1122, "加入组失败")
		}
		return &dto.GroupInfo{
			GroupUUID:   newGroup.UUID,
			GroupName:   newGroup.Name,
			MemberCount: newGroup.MemberCount,
		}, nil
	}
	return nil, cerror.NewCodeError(1122, "组名已存在")
}

// ✅
func (g *GroupService) JoinGroup(req *dto.JoinGroupRequest) (dto.GroupInfo, *cerror.CodeError) {
	group := &model.Group{}
	err := g.DB.Table("group").Where("name = ?", req.GroupName).First(&group).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.GroupInfo{}, cerror.NewCodeError(1122, "组不存在")
	}
	// 检查用户是否已经加入组
	var userGroup model.GroupMember
	err = g.DB.Table("group_member").Where("user_id = ? AND group_id = ? AND delete_at IS NULL", g.UserID, group.Id).First(&userGroup).Error
	if err == nil && userGroup.Id != 0 {
		return dto.GroupInfo{}, cerror.NewCodeError(1122, "用户已经加入组")
	}
	// 加入组
	err = g.DB.Table("group_member").Create(&model.GroupMember{
		GroupId:   group.Id,
		GroupUUID: group.UUID,
		UserId:    g.UserID,
		UserUUID:  g.UserUUID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}).Error

	if err != nil {
		return dto.GroupInfo{}, cerror.NewCodeError(1122, "加入组失败")
	}
	// 更新组的成员数量
	err = g.DB.Table("group").Where("id = ?", group.Id).Update("member_count", group.MemberCount+1).Error
	if err != nil {
		return dto.GroupInfo{}, cerror.NewCodeError(1122, "更新组的成员数量失败")
	}
	return dto.GroupInfo{
		GroupUUID:   group.UUID,
		GroupName:   group.Name,
		MemberCount: group.MemberCount,
	}, nil
}

// ✅ LeaveGroup 离开群组，如果是群主则解散群组
func (g *GroupService) LeaveGroup(req *dto.LeaveGroupRequest) *cerror.CodeError {
	// 查找群组
	group := model.Group{}
	err := g.DB.Table("group").Where("name = ? OR uuid = ? AND delete_at IS NULL", req.GroupName, req.GroupUUID).First(&group).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cerror.NewCodeError(1122, "群组不存在")
	}
	if err != nil {
		return cerror.NewCodeError(1122, "查询群组失败")
	}
	// 检查用户是否在群组中
	var groupMember model.GroupMember
	err = dao.DB.Table("group_member").Where("user_id = ? AND group_id = ? AND delete_at IS NULL", g.UserID, group.Id).First(&groupMember).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cerror.NewCodeError(1122, "您不在该群组中")
	}
	if err != nil {
		return cerror.NewCodeError(1122, "查询群成员失败")
	}
	// 检查是否为群主
	if group.OwnerId == g.UserID {
		// 群主离开，解散群组
		return g.dissolveGroup(group.Id)

	} else {
		// 普通成员离开群组
		return g.leaveMemberGroup(group.Id, group.MemberCount)
	}
}

// ✅ dissolveGroup 解散群组
func (g *GroupService) dissolveGroup(groupId int) *cerror.CodeError {
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
		return cerror.NewCodeError(1122, "删除群成员失败")
	}
	// 删除群组
	err = tx.Table("group").Where("id = ?", groupId).Update("delete_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return cerror.NewCodeError(1122, "删除群组失败")
	}
	// 查询操作用户
	user := model.Users{}
	err = tx.Table("users").Where("id = ?", g.UserID).First(&user).Error

	if err != nil {
		tx.Rollback()
		return cerror.NewCodeError(1122, "查询用户失败")
	}
	// 更新自己的群组操作状态，用于Redis
	tx.Table("users").Where("id = ?", g.UserID).Update("group_version", user.GroupVersion+1)
	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return cerror.NewCodeError(1122, "提交事务失败")
	}
	return nil
}

// ✅leaveMemberGroup 普通成员离开群组
func (g *GroupService) leaveMemberGroup(groupId int, memberCount int) *cerror.CodeError {
	// 开启事务
	tx := dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 删除群成员记录
	err := tx.Table("group_member").Where("user_id = ? AND group_id = ?", g.UserID, groupId).Update("delete_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return cerror.NewCodeError(1122, "离开群组失败")
	}
	// 更新群组成员数量
	newMemberCount := memberCount - 1
	if newMemberCount < 0 {
		newMemberCount = 0
	}
	err = tx.Table("group").Where("id = ?", groupId).Update("member_count", newMemberCount).Error
	if err != nil {
		tx.Rollback()
		return cerror.NewCodeError(1122, "更新群组成员数量失败")
	}
	// 查询操作用户
	user := model.Users{}
	err = tx.Table("users").Where("id = ?", g.UserID).First(&user).Error
	if err != nil {
		tx.Rollback()
		return cerror.NewCodeError(1122, "查询用户失败")
	}
	// 更新自己的群组操作状态，用于Redis
	tx.Table("users").Where("id = ?", g.UserID).Update("group_version", user.GroupVersion+1)
	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return cerror.NewCodeError(1122, "提交事务失败")
	}
	return nil
}

func GetGroupMember(groupUUID string) ([]string, error) {
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

// ✅ GetGroupList 获取自己加入的群组
func (g *GroupService) GetGroupList() ([]dto.GroupInfo, *cerror.CodeError) {
	// 联合查询Group表和GroupMember表
	var groups []dto.GroupInfo
	err := g.DB.Table("group").
		Select("group.uuid as group_uuid, group.name as group_name, group.member_count as member_count").
		Joins("JOIN group_member ON group.uuid = group_member.group_uuid").
		Where("group_member.user_id = ? AND group_member.delete_at IS NULL", g.UserID).
		Where("group.delete_at IS NULL").
		Scan(&groups).Error
	if err != nil {
		return nil, cerror.NewCodeError(1122, "查询群组失败")
	}
	return groups, nil
}
