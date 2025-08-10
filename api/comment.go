package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"
	"fmt"

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
	c.JSON(200, gin.H{"message": "评论成功"})
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
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "点赞成功"})
}
