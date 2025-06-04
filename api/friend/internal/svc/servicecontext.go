package svc

import (
	"chatLion/api/friend/internal/config"
	"chatLion/rpc/friend/friendclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	FriendRPC friendclient.Friend
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		FriendRPC: friendclient.NewFriend(zrpc.MustNewClient(c.FriendRPC)),
	}
}
