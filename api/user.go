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
		c.JSON(400, dto.GetUserInfoResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1019,
			UserInfo: dto.UserInfo{
				Email:    "",
				Username: "",
				Nickname: "",
				Avatar:   "",
			},
		})
		return
	}
	userInfo, err := service.GetUserInfor(uuid)
	if err != nil {
		c.JSON(http.StatusOK, dto.GetUserInfoResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1019,
			UserInfo: dto.UserInfo{
				Email:    "",
				Username: "",
				Nickname: "",
				Avatar:   "",
			},
		})
		return
	}
	c.JSON(http.StatusOK, dto.GetUserInfoResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code:     2004,
		UserInfo: *userInfo,
	})
}
func UpdateUserInfor(c *gin.Context) {
	// 获取当前用户UUID
	uuid := c.GetString("userUuid")
	if uuid == "" {
		c.JSON(http.StatusOK, dto.UpdateUserResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1019,
			Msg:  "token失效",
		})
		return
	}

	// 绑定请求参数
	req := dto.UpdateUserReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.UpdateUserResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1001,
			Msg:  "json格式错误: ",
		})
		return
	}

	// 调用服务层更新用户信息
	err := service.UpdateUserInfor(uuid, &req)
	if err != nil {
		c.JSON(400, dto.UpdateUserResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1027,
			Msg:  "更新用户信息失败",
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, dto.UpdateUserResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "用户信息更新成功",
	})
}
