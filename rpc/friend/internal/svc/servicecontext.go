package svc

import (
	"chatLion/database"
	"chatLion/rpc/friend/internal/config"

	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	Db     *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Db:     database.InitGorm("root:0220059cyCY@tcp(127.0.0.1:3306)/chatLion?charset=utf8mb4&parseTime=True&loc=Local"),
	}
}
