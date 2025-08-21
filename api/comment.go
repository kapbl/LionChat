package api

import (
	"cchat/internal/dao"
	"cchat/internal/dto"
	"cchat/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 创建一条评论
func CreateComment(c *gin.Context) {
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, dto.CreateCommentResponse{
			Code: 1,
			Msg:  err.Error(),
		})
		return
	}
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	username := c.GetString("username")
	uuid := c.GetString("userUuid")
	commentService := service.NewCommentService(dao.DB, username, uuid, int64(iuserId))
	if err := commentService.CreateComment(&req); err != nil {
		c.JSON(400, dto.CreateCommentResponse{
			Code: int(err.Code),
			Msg:  err.Msg,
		})
		return
	}
	c.JSON(200, dto.CreateCommentResponse{
		Code: 0,
		Msg:  "评论成功",
	})
}

// 点赞动态
func LikeComment(c *gin.Context) {
	var req dto.LikeCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, dto.LikeCommentResponse{
			Code: 1,
			Msg:  err.Error(),
		})
		return
	}
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	commentService := service.NewCommentService(dao.DB, username, uuid, int64(iuserId))
	if err := commentService.LikeComment(req.MomentID); err != nil {
		c.JSON(200, dto.LikeCommentResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: int(err.Code),
			Msg:  err.Msg,
		})
		return
	}
	c.JSON(200, dto.LikeCommentResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "点赞成功",
	})
}

// 获取动态下的评论列表
func GetCommentList(c *gin.Context) {
	moment_id := c.Query("moment_id")
	i_moment_id, err := strconv.Atoi(moment_id)
	if err != nil {
		c.JSON(200, dto.GetCommentList{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "动态id出错",
		})
	}
	userId, _ := c.Get("userId")
	iuserId := userId.(int)
	uuid := c.GetString("userUuid")
	username := c.GetString("username")
	commentService := service.NewCommentService(dao.DB, username, uuid, int64(iuserId))
	resp, err2 := commentService.GetCommentsList(int64(i_moment_id))
	if err2 != nil {
		c.JSON(200, dto.GetCommentList{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: int(err2.Code),
			Msg:  err2.Msg,
		})
		return
	}
	c.JSON(200, dto.GetCommentList{
		BaseResponse: dto.BaseResponse{},
		Code:         0,
		Msg:          "获取成功",
		CommentList:  resp,
	})
}
