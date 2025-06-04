package svc

import (
	"chatLion/api/group/internal/config"
	"chatLion/rpc/group/groupclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config   config.Config
	GroupRPC groupclient.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		GroupRPC: groupclient.NewGroup(zrpc.MustNewClient(c.GroupRPC)),
	}
}
