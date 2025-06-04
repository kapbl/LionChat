package logic

import (
	"context"
	"errors"

	"chatLion/model"
	"chatLion/rpc/group/group"
	"chatLion/rpc/group/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateGroupLogic) CreateGroup(in *group.CreateGroupRequest) (*group.CreateGroupResponse, error) {
	// todo: add your logic here and delete this line
	var currentGroup model.Group
	err := l.svcCtx.Db.Take(&currentGroup, "name = ?", in.CreateGroupName).Error
	if err == nil {
		return &group.CreateGroupResponse{
			CreateMessage: "组已经存在",
		}, errors.New("group exiting")
	}
	// 创建群组
	// 自动将创建者加入到这个群组中
	var currentUser model.UserModel
	err = l.svcCtx.Db.Take(&currentUser, "email = ?", in.CreateOwnerName).Error
	if err != nil {
		return &group.CreateGroupResponse{
			CreateMessage: "创建者不存在",
		}, errors.New("invaild creater")
	}
	currentGroup.Name = in.CreateGroupName
	currentGroup.Description = in.CreateGroupDescription
	currentGroup.HostID = currentUser.ID
	currentGroup.Members = make([]model.UserModel, 0)
	currentGroup.Members = append(currentGroup.Members, currentUser)
	// !创建这个群组
	err = l.svcCtx.Db.Create(&currentGroup).Error
	if err != nil {
		return &group.CreateGroupResponse{
			CreateMessage: "组创建失败",
		}, errors.New("group create failed")
	}

	return &group.CreateGroupResponse{
		CreateMessage: "组创建成功",
	}, nil
}
