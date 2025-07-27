package router

import (
	"cchat/api"
	"cchat/internal/middlewares"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var webEngine *gin.Engine

// var AppRouters map[string]func(c *gin.Context)
var AppRouterGroups map[string]*gin.RouterGroup = map[string]*gin.RouterGroup{}

func InitWebEngine() {
	// 初始化路由
	gin.SetMode(gin.ReleaseMode)
	webEngine = gin.Default()
	InitCors()
	InitRouterGroups()
	InitMiddleware()
	InitRouter()
}

func RunEngine() {
	// 启动路由
	webEngine.Run(":8081")
}

func InitRouterGroups() {
	// 用户相关的路由组
	AppRouterGroups["user"] = webEngine.Group("v1/api/user")
	AppRouterGroups["webSocket"] = webEngine.Group("v1/api/webSocket")
	AppRouterGroups["friend"] = webEngine.Group("v1/api/friend")
	AppRouterGroups["message"] = webEngine.Group("v1/api/message")
	AppRouterGroups["group"] = webEngine.Group("v1/api/group")

	log.Printf("初始化了路由组: %d 个", len(AppRouterGroups))
}

func InitCors() {
	webEngine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},           // 允许所有域名
		AllowMethods:     []string{"GET", "POST"}, // 允许的方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
}
func InitMiddleware() {
	// AppRouterGroups["webSocket"].Use(middlewares.JwtMiddleware())
	AppRouterGroups["friend"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["group"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
}
func InitRouter() {
	// *** 用户相关的路由
	// todo 用户获取
	AppRouterGroups["user"].POST("/login", api.Login)
	AppRouterGroups["user"].POST("/register", api.Register)
	// todo 用户登出
	AppRouterGroups["user"].POST("/logout", api.Logout)
	// AppRouterGroups["user"].GET("/getUserInfo", api.GetUserInfo)
	// AppRouterGroups["user"].POST("/updateprofile", api.UpdateProfile)
	// AppRouterGroups["user"].POST("/changePassword", api.ChangePassword)

	// ***  socket相关的路由
	AppRouterGroups["webSocket"].GET("/connect", api.WebSocketConnect)
	//*** 好友相关的路由
	AppRouterGroups["friend"].GET("/search", api.SearchClientByUserName)
	AppRouterGroups["friend"].POST("/addFriend", api.AddSearchClientByUserName)
	AppRouterGroups["friend"].POST("/handleRequest", api.ReceiveFriendRequest)
	AppRouterGroups["friend"].POST("/handleResponse", api.HandleFriendRequest)
	AppRouterGroups["friend"].GET("/getFriendList", api.GetFriendList)
	//*** 消息相关的路由
	AppRouterGroups["message"].GET("/getUnreadMessage", api.GetUnreadMessage)
	//*** 群组相关的路由
	AppRouterGroups["group"].POST("/createGroup", api.CreateGroup)
	AppRouterGroups["group"].POST("/joinGroup", api.JoinGroup)
	AppRouterGroups["group"].POST("/leaveGroup", api.LeaveGroup)
	AppRouterGroups["group"].GET("/getGroupList", api.GetGroupList)
	// AppRouterGroups["group"].POST("/changeGroupInfo", api.ChangeGroupInfo)
	// AppRouterGroups["group"].POST("/changeGroupAvatar", api.ChangeGroupAvatar)
	// AppRouterGroups["group"].POST("/changeGroupOwner", api.ChangeGroupOwner)
}
