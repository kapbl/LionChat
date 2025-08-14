package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	req := dto.RegisterRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.RegisterResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("request_id"),
			},
			Code: 1001,
			Msg:  "请求参数错误",
		})
		return
	}
	err := service.Register(&req)
	if err != nil {
		fmt.Println(c.GetString("request_id"))
		c.JSON(http.StatusOK, dto.RegisterResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("request_id"),
			},
			Code: err.Code,
			Msg:  err.Msg,
		})
		return
	}
	c.JSON(http.StatusOK, dto.RegisterResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("request_id"),
		},
		Code: 2000,
		Msg:  "注册成功",
	})
}
