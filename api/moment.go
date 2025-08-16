package api

import (
	"cchat/internal/dao"
	"cchat/internal/dto"
	"cchat/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateMoment(c *gin.Context) {
	// 解析jwt
	uuid := c.GetString("userUuid")
	if uuid == "" {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "用户未登录",
		})
		return
	}

	// 解析请求体
	var moment dto.MomentCreateRequest
	if err := c.ShouldBindJSON(&moment); err != nil {
		c.JSON(http.StatusOK, dto.MomentCreateResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "请求参数错误",
		})
		return
	}
	// 校验moment
	if moment.Content == "" {
		c.JSON(http.StatusOK, dto.MomentCreateResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "动态内容不能为空",
		})
		return
	}
	userID, exist := c.Get("userId")
	if !exist {
		c.JSON(http.StatusOK, dto.MomentCreateResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "用户ID不存在",
		})
		return
	}
	iuserID := userID.(int)
	momentService := service.NewMomentService(int64(iuserID), uuid, dao.DB)
	// 创建动态
	err2 := momentService.CreateMoment(&moment)
	if err2 != nil {
		c.JSON(http.StatusOK, dto.MomentCreateResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "创建动态失败",
		})
		return
	}
	// 返回动态ID
	c.JSON(http.StatusOK, dto.MomentCreateResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "创建动态成功",
	})
}

// 获取动态列表
func ListMoment(c *gin.Context) {
	// 解析jwt
	uuid := c.GetString("userUuid")
	if uuid == "" {
		c.JSON(http.StatusOK, dto.MomentListResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "用户未登录",
		})
		return
	}
	userID, exist := c.Get("userId")
	if !exist {
		c.JSON(http.StatusOK, dto.MomentCreateResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "用户ID不存在",
		})
		return
	}
	iUseID := userID.(int)
	// 获取动态列表
	momentService := service.NewMomentService(int64(iUseID), uuid, dao.DB)
	momentList, err := momentService.ListMoment()

	if err != nil {
		c.JSON(http.StatusOK, dto.MomentListResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "获取动态列表失败",
		})
		return
	}

	// 返回动态列表
	c.JSON(http.StatusOK, dto.MomentListResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "获取动态列表成功",
		Data: momentList,
	})
}

// GetCommentsByMomentID 根据动态ID获取评论列表
func GetCommentsByMomentID(c *gin.Context) {
	// 获取动态ID参数
	momentIDStr := c.Param("momentId")
	if momentIDStr == "" {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "动态ID不能为空",
		})
		return
	}

	// 转换动态ID为int64
	momentID, err := strconv.ParseInt(momentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "动态ID格式错误",
		})
		return
	}

	// 调用服务层获取评论列表
	comments, err := service.GetCommentsByMomentID(momentID)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "获取评论列表失败: " + err.Error(),
		})
		return
	}

	// 返回评论列表
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: comments,
	})
}
