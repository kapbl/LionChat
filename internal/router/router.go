package router

import (
	"cchat/api"
	"cchat/internal/middlewares"
	"cchat/pkg/config"
	"cchat/pkg/logger"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	logger.Info("Web引擎初始化完成")
}

func RunEngine(c *config.Config) {
	server := &http.Server{
		Addr:         "localhost:" + strconv.Itoa(c.Server.Port),
		Handler:      webEngine,
		ReadTimeout:  time.Duration(c.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.Server.WriteTimeout) * time.Second,
	}

	logger.Info("启动Web服务器", zap.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("启动Web服务器失败", zap.Error(err))
	}
}

func InitRouterGroups() {
	// 用户相关的路由组
	AppRouterGroups["user"] = webEngine.Group("v1/api/user")
	AppRouterGroups["webSocket"] = webEngine.Group("v1/api/webSocket")
	AppRouterGroups["friend"] = webEngine.Group("v1/api/friend")
	AppRouterGroups["message"] = webEngine.Group("v1/api/message")
	AppRouterGroups["group"] = webEngine.Group("v1/api/group")
	AppRouterGroups["monitor"] = webEngine.Group("v1/api/monitor")
	AppRouterGroups["profile"] = webEngine.Group("v1/api/profile")

	logger.Info("路由组初始化完成",
		zap.Int("groupCount", len(AppRouterGroups)),
		zap.Strings("groups", getRouterGroupNames()))
}

func getRouterGroupNames() []string {
	names := make([]string, 0, len(AppRouterGroups))
	for name := range AppRouterGroups {
		names = append(names, name)
	}
	return names
}

func InitCors() {
	webEngine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},           // 允许所有域名
		AllowMethods:     []string{"GET", "POST"}, // 允许的方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	logger.Info("CORS中间件配置完成")
}

func InitMiddleware() {
	// AppRouterGroups["webSocket"].Use(middlewares.JwtMiddleware())
	AppRouterGroups["friend"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["group"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["profile"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	// AppRouterGroups["monitor"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	logger.Info("JWT中间件配置完成",
		zap.Strings("protected_groups", []string{"friend", "group", "monitor"}))
}

func InitRouter() {
	// *** 用户相关的路由
	AppRouterGroups["user"].POST("/login", api.Login)
	AppRouterGroups["user"].POST("/register", api.Register)
	AppRouterGroups["user"].POST("/logout", api.Logout)
	AppRouterGroups["profile"].GET("/getProfileInfo", api.GetUserInfor)
	AppRouterGroups["profile"].PUT("/updateProfile", api.UpdateUserInfor)

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
	// todo
	AppRouterGroups["group"].GET("/getGroupInfo", api.GetGroupInfo)
	// todo
	AppRouterGroups["group"].PUT("/changeGroupInfo", api.ChangeGroupInfo)

	// *** 监控相关的路由
	AppRouterGroups["monitor"].GET("/clients", api.GetClients)
	AppRouterGroups["monitor"].GET("/serverInfo", api.GetServerInfor)

	logger.Info("API路由注册完成")
}
