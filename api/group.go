package api

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
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
	resp, err := groupService.CreateGroup(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Base{
			Code: 1,
			Data: "创建组失败",
		})
		return
	}
	c.JSON(http.StatusOK, dto.Base{
		Code: 0,
		Data: resp,
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
	// 获取用户ID
	userUuid, exists := c.Get("userUuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 绑定请求参数
	var req dto.LeaveGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 参数验证
	if req.GroupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "群组名称不能为空"})
		return
	}

	// 获取用户信息
	var user model.Users
	err := dao.DB.Table("users").Where("uuid = ?", userUuid).First(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	// 调用service层离开群组
	groupService := &service.GroupService{UserId: int(user.Id)}
	resp, err := groupService.LeaveGroup(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": resp.Message,
		"data":    resp,
	})
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

func GetGroupInfo(c *gin.Context) {

}

func ChangeGroupInfo(c *gin.Context) {

}
