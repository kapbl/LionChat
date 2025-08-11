package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateComment(c *gin.Context) {
	var req dto.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	userId := c.GetInt("userId")
	if err := service.CreateComment(int64(userId), req.MomentID, req.Content); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, dto.Base{
		Code: 0,
		Data: "发布评论成功",
	})
}

func LikeComment(c *gin.Context) {
	var req dto.LikeCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	userId := c.GetInt("userId")
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	fmt.Println(userId, uuid, username)

	if err := service.LikeComment(int64(userId), req.MomentID); err != nil {
		c.JSON(400, dto.Base{
			Code: 1,
			Data: err.Error(),
		})
		return
	}
	c.JSON(400, dto.Base{
		Code: 0,
		Data: "点赞成功",
	})
}

func GetCommentList(c *gin.Context) {
	moment_id := c.Query("moment_id")
	i_moment_id, err := strconv.Atoi(moment_id)
	if err != nil {
		c.JSON(200, dto.Base{
			Code: 1,
			Data: "动态id出错",
		})
	}
	resp, err := service.GetCommentsByMomentID(int64(i_moment_id))
	if err != nil {
		c.JSON(200, dto.Base{
			Code: 1,
			Data: err.Error(),
		})
	}
	c.JSON(200, dto.Base{
		Code: 0,
		Data: resp,
	})
}
