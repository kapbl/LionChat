package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	req := dto.LoginReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.Base{
			Data: "格式错误",
		})
		return
	}
	currentUser, err := service.Login(&req)
	if err != nil {
		c.JSON(400, dto.Base{
			Data: err,
		})
	}
	c.JSON(200, currentUser)
}
