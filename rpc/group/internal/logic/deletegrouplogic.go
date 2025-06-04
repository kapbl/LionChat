package logic

import (
	"context"
	"errors"
	"log"

	"chatLion/model"
	"chatLion/rpc/group/group"
	"chatLion/rpc/group/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGroupLogic {
	return &DeleteGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteGroupLogic) DeleteGroup(in *group.DeleteGroupRequest) (*group.DeleteGroupResponse, error) {
	// todo: add your logic here and delete this line
	targetGroup := new(model.Group)
	err := l.svcCtx.Db.First(targetGroup, "name = ?", in.DeleteGroupName).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &group.DeleteGroupResponse{
			DeleteMessage: "未查询到这个群组",
		}, err
	}
	currUser := new(model.UserModel)
	err = l.svcCtx.Db.First(currUser, "email = ?", in.UserName).Error
	if err != nil {
		return &group.DeleteGroupResponse{
			DeleteMessage: "用户不存在",
		}, err
	}
	res := ""
	// 处理房主的情况,直接删除这个群组
	if targetGroup.HostID == currUser.ID {
		l.svcCtx.Db.Select("Members").Delete(&targetGroup)
		res = "成功解散" + targetGroup.Name + "群组"
	} else {
		// 仅仅接触用户和群组的关系
		err := l.svcCtx.Db.Model(targetGroup).Association("Members").Delete(currUser)
		res = "成功退出" + targetGroup.Name + "群组"
		if err != nil {
			log.Println(err)
		}
	}

	return &group.DeleteGroupResponse{
		DeleteMessage: res,
	}, nil
}
