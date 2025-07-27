package api

import (
	"cchat/internal/dto"
	"cchat/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateGroup(c *gin.Context) {
	var req dto.CreateGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "参数错误",
		})
		return
	}
	if req.GroupName == "" {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "组名不能为空",
		})
	}

	userId := c.GetInt("userId")        // 自己的id
	uuid := c.GetString("userUuid")     // 自己的uuid
	username := c.GetString("username") // 自己的账户名字
	groupService := service.GroupService{
		UserId:   userId,
		UserUUID: uuid,
		Username: username,
	}
	err := groupService.CreateGroup(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "创建组失败",
		})
		return
	}
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: "创建组成功",
	})
}

func JoinGroup(c *gin.Context) {
	req := dto.JoinGroupReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "参数错误",
		})
		return
	}
	if req.GroupName == "" {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "参数错误",
		})
	}
	userId := c.GetInt("userId")        // 自己的id
	uuid := c.GetString("userUuid")     // 自己的uuid
	username := c.GetString("username") // 自己的账户名字
	groupService := service.GroupService{
		UserId:   userId,
		UserUUID: uuid,
		Username: username,
	}
	res, err := groupService.JoinGroup(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "加入组失败",
		})
	}
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: res,
	})
}

func LeaveGroup(c *gin.Context) {

}

func GetGroupList(c *gin.Context) {
	userId := c.GetInt("userId")
	groupService := service.GroupService{
		UserId: userId,
	}
	res, err := groupService.GetGroupList()
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "获取组列表失败",
		})
	}
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: res,
	})
}
