package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	req := dto.LoginRequestDTO{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.LoginResponseDTO{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("request_id"),
			},
			Code: 1000, // json 格式错误
			Msg:  "格式错误",
			Data: dto.LoginData{},
		})
		return
	}
	currentUser, err := service.Login(&req)
	if err != nil {
		c.JSON(400, dto.LoginResponseDTO{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("request_id"),
			},
			Code: 1001, // 登录失败
			Msg:  "登录失败",
			Data: dto.LoginData{},
		})
		return
	}
	c.JSON(200, dto.LoginResponseDTO{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("request_id"),
		},
		Code: 0,
		Msg:  "登录成功",
		Data: currentUser,
	})
}
