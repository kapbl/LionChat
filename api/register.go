package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Base{
			Code: 400,
			Data: "请求参数错误",
		})
		return
	}
	_, err := service.Register(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 500,
			Data: err.Error(),
		})
	}
	c.JSON(http.StatusOK, dto.Base{
		Code: 200,
		Data: "注册成功",
	})
}
