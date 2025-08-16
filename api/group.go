package api

import (
	"cchat/internal/dao"
	"cchat/internal/dto"
	"cchat/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 创建一个群组
func CreateGroup(c *gin.Context) {
	req := dto.CreateGroupRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.CreateGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1122,
			Msg:  "参数错误",
			GroupInfo: dto.GroupInfo{
				GroupUUID:   "",
				GroupName:   "",
				MemberCount: 0,
			},
		})
		return
	}
	if req.GroupName == "" {
		c.JSON(http.StatusOK, dto.CreateGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1333,
			Msg:  "组名不能为空",
			GroupInfo: dto.GroupInfo{
				GroupUUID:   "",
				GroupName:   "",
				MemberCount: 0,
			},
		})
		return
	}
	userId := c.GetInt("userId")
	userUUID := c.GetString("userUuid")
	groupService := service.NewGroupService(userId, userUUID, dao.DB)
	resp, err := groupService.CreateGroup(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.CreateGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "创建组失败",
		})
		return
	}
	c.JSON(http.StatusOK, dto.CreateGroupResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "创建成功",
		GroupInfo: dto.GroupInfo{
			GroupUUID:   resp.GroupUUID,
			GroupName:   resp.GroupName,
			MemberCount: resp.MemberCount,
		},
	})
}

// 加入群组
func JoinGroup(c *gin.Context) {
	req := dto.JoinGroupRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.JoinGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1122,
			Msg:  "参数错误",
			GroupInfo: dto.GroupInfo{
				GroupUUID:   "",
				GroupName:   "",
				MemberCount: 0,
			},
		})
		return
	}
	// 只能通过组名或组uuid加入群组
	if req.GroupName == "" && req.GroupUUID == "" {
		c.JSON(http.StatusOK, dto.JoinGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1122,
			Msg:  "参数错误",
			GroupInfo: dto.GroupInfo{
				GroupUUID:   "",
				GroupName:   "",
				MemberCount: 0,
			},
		})
	}
	userId := c.GetInt("userId")    // 自己的id
	uuid := c.GetString("userUuid") // 自己的uuid
	groupService := service.NewGroupService(userId, uuid, dao.DB)
	res, err := groupService.JoinGroup(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.JoinGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1122,
			Msg:  "加入组失败",
			GroupInfo: dto.GroupInfo{
				GroupUUID:   "",
				GroupName:   "",
				MemberCount: 0,
			},
		})
	}
	c.JSON(http.StatusOK, dto.JoinGroupResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "加入成功",
		GroupInfo: dto.GroupInfo{
			GroupUUID:   res.GroupUUID,
			GroupName:   res.GroupName,
			MemberCount: res.MemberCount,
		},
	})
}

func LeaveGroup(c *gin.Context) {
	// 获取用户ID
	userID := c.GetInt("userId")
	userUUID := c.GetString("userUuid")
	GroupUUID := c.Query("groupUuid")
	GroupName := c.Query("groupName")
	// 参数验证
	if GroupUUID == "" && GroupName == "" {
		c.JSON(http.StatusBadRequest, dto.LeaveGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1122,
			Msg:  "参数错误",
		})
		return
	}
	groupService := service.NewGroupService(userID, userUUID, dao.DB)
	req := dto.LeaveGroupRequest{
		GroupUUID: GroupUUID,
		GroupName: GroupName,
	}
	// 调用service层离开群组
	err := groupService.LeaveGroup(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.LeaveGroupResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1122,
			Msg:  "参数错误",
		})
		return
	}

	c.JSON(200, dto.LeaveGroupResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 1122,
		Msg:  "离开组成功",
	})
}

// 获取自己的好友列表
func GetGroupList(c *gin.Context) {
	userId := c.GetInt("userId")
	userUUID := c.GetString("userUuid")
	groupService := service.NewGroupService(userId, userUUID, dao.DB)
	res, err := groupService.GetGroupList()
	if err != nil {
		c.JSON(http.StatusOK, dto.GetGroupsResponse{
			BaseResponse: dto.BaseResponse{
				RequestID: c.GetString("requestId"),
			},
			Code: 1,
			Msg:  "获取组列表失败",
		})
	}
	c.JSON(http.StatusOK, dto.GetGroupsResponse{
		BaseResponse: dto.BaseResponse{
			RequestID: c.GetString("requestId"),
		},
		Code: 0,
		Msg:  "获取组列表成功",
		Data: res,
	})
}

func GetGroupInfo(c *gin.Context) {

}

func ChangeGroupInfo(c *gin.Context) {

}
