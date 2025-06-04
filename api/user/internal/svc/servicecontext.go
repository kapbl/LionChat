package svc

import (
	"chatLion/api/user/internal/config"
	"chatLion/rpc/user/userclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	UserRPC userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		UserRPC: userclient.NewUser(zrpc.MustNewClient(c.UserRPC)),
	}
}
