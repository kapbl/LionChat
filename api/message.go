package api

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UnreadMessageResponse 未读消息响应结构
type UnreadMessageResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *UnreadMessageData `json:"data,omitempty"`
}

type UnreadMessageData struct {
	Messages []MessageInfo `json:"messages"`
	Total    int           `json:"total"`
}

type MessageInfo struct {
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id"`
	ReceiveID string `json:"receive_id"`
	Content   string `json:"content"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

// GetUnreadMessage 获取用户的未读消息
func GetUnreadMessage(c *gin.Context) {
	// 获取用户UUID参数
	userUUID := c.GetString("userUuid")
	if userUUID == "" {
		c.JSON(http.StatusBadRequest, UnreadMessageResponse{
			Code:    400,
			Message: "用户UUID不能为空",
		})
		return
	}

	// 获取分页参数
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 查询未读消息
	var messages []model.Message
	err := dao.DB.Table("message").
		Where("receive_id = ? AND status = 0", userUUID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, UnreadMessageResponse{
			Code:    500,
			Message: "查询未读消息失败: " + err.Error(),
		})
		return
	}

	// 查询总数
	var total int64
	err = dao.DB.Table("message").
		Where("receive_id = ? AND status = 0", userUUID).
		Count(&total).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, UnreadMessageResponse{
			Code:    500,
			Message: "查询未读消息总数失败: " + err.Error(),
		})
		return
	}

	// 转换为响应格式
	messageInfos := make([]MessageInfo, len(messages))
	for i, msg := range messages {
		messageInfos[i] = MessageInfo{
			MessageID: msg.MessageID,
			SenderID:  msg.SenderID,
			ReceiveID: msg.ReceiveID,
			Content:   msg.Content,
			Status:    int(msg.Status),
			CreatedAt: msg.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// 返回成功响应
	c.JSON(http.StatusOK, UnreadMessageResponse{
		Code:    200,
		Message: "获取未读消息成功",
		Data: &UnreadMessageData{
			Messages: messageInfos,
			Total:    int(total),
		},
	})
}

// MarkMessageAsReadRequest 标记消息已读请求结构
type MarkMessageAsReadRequest struct {
	MessageIDs []string `json:"message_ids" binding:"required"`
}

// MarkMessageAsReadResponse 标记消息已读响应结构
type MarkMessageAsReadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		UpdatedCount int `json:"updated_count"`
	} `json:"data,omitempty"`
}

// MarkMessageAsRead 标记消息为已读
func MarkMessageAsRead(c *gin.Context) {
	var req MarkMessageAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MarkMessageAsReadResponse{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	if len(req.MessageIDs) == 0 {
		c.JSON(http.StatusBadRequest, MarkMessageAsReadResponse{
			Code:    400,
			Message: "消息ID列表不能为空",
		})
		return
	}
	// 获取用户UUID参数
	userUUID := c.GetString("userUuid")
	if userUUID == "" {
		c.JSON(http.StatusBadRequest, MarkMessageAsReadResponse{
			Code:    400,
			Message: "用户UUID不能为空",
		})
		return
	}
	fmt.Println(req.MessageIDs)

	// 更新消息状态为已读
	result := dao.DB.Table("message").
		Where("receive_id = ? AND message_id IN ? AND status = 0", userUUID, req.MessageIDs).
		Update("status", 1)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, MarkMessageAsReadResponse{
			Code:    500,
			Message: "标记消息已读失败: " + result.Error.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, MarkMessageAsReadResponse{
		Code:    200,
		Message: "标记消息已读成功",
		Data: &struct {
			UpdatedCount int `json:"updated_count"`
		}{
			UpdatedCount: int(result.RowsAffected),
		},
	})
}
