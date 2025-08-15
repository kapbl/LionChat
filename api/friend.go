package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"

	"github.com/gin-gonic/gin"
)

// ✅搜索用户：Username, Nickname， Email
func SearchClient(c *gin.Context) {
	information := c.Query("information")
	if information == "" {
		c.JSON(400, dto.SearchClientResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Data: nil,
		})
		return
	}
	res, err := service.SearchClient(information)

	if err != nil {
		c.JSON(400, dto.SearchClientResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: err.Code,
			Data: nil,
		})
		return
	}
	c.JSON(200, dto.SearchClientResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 200,
		Data: []dto.UserInfo{*res},
	})
}

// 处理好友请求：同意 和 拒绝
func HandleFriendResponse(c *gin.Context) {
	req := dto.HandleFriendRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.HandleFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Msg:  "json err",
		})
		return
	}
	if req.TargetUsername == "" {
		c.JSON(400, dto.HandleFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Msg:  "target_username err",
		})
		return
	}
	if req.Status != 0 && req.Status != 1 {
		c.JSON(400, dto.HandleFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Msg:  "status err",
		})
		return
	}
	// go to service
	err := service.HandleFriendRequest(&req, c.GetInt("userId"))
	if err != nil {
		c.JSON(400, dto.HandleFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: err.Code,
			Msg:  err.Msg,
		})
		return
	}
	c.JSON(200, dto.HandleFriendResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 2200,
		Msg:  "success",
	})
}

// ✅ 加好友请求
func AddFriend(c *gin.Context) {
	req := dto.AddFriendRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "json err")
	}
	if req.TargetUsername == "" {
		c.JSON(400, dto.AddFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Msg:  "target_user_name err",
		})
		return
	}
	if req.Content == "" {
		c.JSON(400, dto.AddFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Msg:  "content err",
		})
		return
	}
	userId := c.GetInt("userId")
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	err := service.AddFriend(&req, userId, uuid, username)
	if err != nil {
		c.JSON(400, dto.AddFriendResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: err.Code,
			Msg:  err.Msg,
		})
		return
	}
	c.JSON(200, dto.AddFriendResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 200,
		Msg:  "success",
	})
}

// 获取好友列表
func GetFriendList(c *gin.Context) {
	uuid := c.GetString("userUuid")
	userId := c.GetInt("userId")
	friendList, err := service.GetFriendList(uuid, userId)
	if err != nil {
		c.JSON(400, dto.Base{
			Data: err.Error(),
		})
		return
	}
	c.JSON(200, dto.Base{
		Code: 200,
		Data: friendList,
	})
}
