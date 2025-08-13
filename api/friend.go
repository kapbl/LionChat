package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"

	"github.com/gin-gonic/gin"
)

func SearchClientByUserName(c *gin.Context) {
	username := c.Query("username")

	if username == "" {
		c.JSON(400, dto.Base{
			Data: "参数错误",
		})
		return
	}
	res, err := service.SearchClientByUserName(username)
	if err != nil {
		c.JSON(404, dto.Base{
			Data: "未查询到" + username,
		})
		return
	}
	c.JSON(200, dto.Base{
		Code: 200,
		Data: res,
	})
}

// 加好友
func AddSearchClientByUserName(c *gin.Context) {
	addReq := dto.AddFriendReq{}
	if err := c.ShouldBindJSON(&addReq); err != nil {
		c.JSON(400, "json err")
	}
	userId := c.GetInt("userId")
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	resp, err := service.AddSearchClientByUserName(&addReq, userId, uuid, username)
	if err != nil {
		c.JSON(400, dto.Base{
			Code: 400,
			Data: err.Error(),
		})
	}
	c.JSON(200, dto.Base{
		Code: 200,
		Data: resp,
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

// 接收陌生人加好友请求，打开软件调用一次
func ReceiveFriendRequest(c *gin.Context) {
	req := dto.HandleFriendRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "json err")
	}
	userId := c.GetInt("userId")        // 自己的id
	uuid := c.GetString("userUuid")     // 自己的uuid
	username := c.GetString("username") // 自己的账户名字
	err := service.ReceiveFriendRequest(&req, userId, uuid, username)
	if err != nil {
		c.JSON(400, dto.Base{
			Data: err.Error(),
		})
		return
	}
	c.JSON(200, dto.Base{
		Data: "同意加好友",
	})
}

// 处理已经发送的好友请求
func HandleFriendRequest(c *gin.Context) {
	req := dto.HandleFriendRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "json err")
	}
	userId := c.GetInt("userId")        // 自己的id
	uuid := c.GetString("userUuid")     // 自己的uuid
	username := c.GetString("username") // 自己的账户名字
	err := service.HandleFriendRequest(&req, userId, uuid, username)
	if err != nil {
		c.JSON(400, dto.Base{
			Data: err.Error(),
		})
		return
	}
	c.JSON(200, dto.Base{
		Data: "同意加好友",
	})
}
