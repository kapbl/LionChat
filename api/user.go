package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserInfor(c *gin.Context) {
	uuid := c.GetString("userUuid")
	if uuid == "" {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "用户不存在",
		})
		return
	}
	userInfo, err := service.GetUserInfor(uuid)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: userInfo,
	})
}
func UpdateUserInfor(c *gin.Context) {
	// 获取当前用户UUID
	uuid := c.GetString("userUuid")
	if uuid == "" {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "用户不存在",
		})
		return
	}

	// 绑定请求参数
	var req service.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 调用服务层更新用户信息
	err := service.UpdateUserInfor(uuid, &req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: "用户信息更新成功",
	})
}
