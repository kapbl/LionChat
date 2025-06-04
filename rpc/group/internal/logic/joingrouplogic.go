package logic

import (
	"context"

	"chatLion/model"
	"chatLion/rpc/group/group"
	"chatLion/rpc/group/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type JoinGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewJoinGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinGroupLogic {
	return &JoinGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *JoinGroupLogic) JoinGroup(in *group.JoinGroupRequest) (*group.JoinGroupResponse, error) {
	// todo: add your logic here and delete this line
	var currentGroup model.Group
	if err := l.svcCtx.Db.Take(&currentGroup, "name = ?", in.JoinGroupName).Error; err != nil {
		return &group.JoinGroupResponse{
			JoinMessage: "组不存在",
		}, nil
	}
	var joiner model.UserModel
	if err := l.svcCtx.Db.Take(&joiner, "email = ?", in.JoinerName).Error; err != nil {
		return &group.JoinGroupResponse{
			JoinMessage: "加入的用户不存在",
		}, nil
	}
	// 关联组与成员
	l.svcCtx.Db.Model(&currentGroup).Association("Members").Append(&joiner)

	return &group.JoinGroupResponse{}, nil
}
