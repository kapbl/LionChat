package api

import (
	"cchat/internal/dao"
	"cchat/internal/dto"
	"cchat/internal/service"

	"github.com/gin-gonic/gin"
)

// ✅搜索用户：Username, Nickname， Email
func SearchClient(c *gin.Context) {
	information := c.Query("information")
	if information == "" {
		c.JSON(200, dto.SearchClientResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 400,
			Data: nil,
		})
		return
	}
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	friendService := service.NewFriendService(int64(iuserId), uuid, username, dao.DB)
	res, err := friendService.SearchClient(information)
	if err != nil {
		c.JSON(200, dto.SearchClientResponse{
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

// ✅ 处理好友请求：同意 和 拒绝
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
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	friendService := service.NewFriendService(int64(iuserId), uuid, username, dao.DB)
	err := friendService.HandleFriendRequest(&req)
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
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	friendService := service.NewFriendService(int64(iuserId), uuid, username, dao.DB)
	err := friendService.AddFriend(&req)
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
	// 获取userID用户的所有好友-列表
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	friendService := service.NewFriendService(int64(iuserId), uuid, username, dao.DB)
	friendList, err := friendService.GetFriendList()
	if err != nil {
		c.JSON(200, dto.FriendListResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: err.Code,
			Msg:  err.Msg,
		})
		return
	}
	c.JSON(200, dto.FriendListResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 200,
		Data: friendList,
	})
}
