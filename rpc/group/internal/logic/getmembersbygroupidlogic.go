package logic

import (
	"context"
	"log"

	"chatLion/model"
	"chatLion/rpc/group/group"
	"chatLion/rpc/group/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMembersByGroupIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMembersByGroupIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMembersByGroupIDLogic {
	return &GetMembersByGroupIDLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMembersByGroupIDLogic) GetMembersByGroupID(in *group.GetMembersRequest) (*group.GetMembersResponse, error) {
	// 在数据库中查找这个群组的所有成员
	var currentGroup model.Group
	err := l.svcCtx.Db.Preload("Members").First(&currentGroup, "name = ?", in.GroupId).Error
	if err != nil {
		log.Println(err)
	}
	members := make([]string, len(currentGroup.Members))
	for i, member := range currentGroup.Members {
		members[i] = member.Email
	}
	return &group.GetMembersResponse{Members: members}, nil
}
