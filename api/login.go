package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	req := dto.LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.LoginResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("request_id"),
			},
			Code:        1000,
			AccessToken: "",
		})
		return
	}
	accessToken, err := service.Login(&req)
	if err != nil {
		c.JSON(400, dto.LoginResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("request_id"),
			},
			Code:        err.Code,
			AccessToken: "",
		})
		return
	}
	c.JSON(200, dto.LoginResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("request_id"),
		},
		Code:        2002,
		AccessToken: accessToken,
	})
}
