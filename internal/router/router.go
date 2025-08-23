package router

import (
	"cchat/api"
	"cchat/internal/middlewares"
	"cchat/pkg/config"
	"cchat/pkg/logger"
	"io"
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

func InitWebEngine(c *config.Config) {
	// 初始化路由
	gin.SetMode(gin.ReleaseMode)
	// 禁用Gin的控制台输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	// 如果是生产环境，完全禁用日志输出
	if c.Server.Environment == "prod" {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
	}
	webEngine = gin.New()
	// 只在开发环境使用Recovery中间件
	if c.Server.Environment == "dev" || c.Server.Environment == "development" {
		webEngine.Use(gin.Recovery())
	}
	InitCors(c)
	InitRouterGroups()
	InitMiddleware()
	InitRouter()
	logger.Info("Web引擎初始化完成")
}

func RunEngine(c *config.Config) {
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(c.Server.Port),
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
	AppRouterGroups["moment"] = webEngine.Group("v1/api/moment")
	AppRouterGroups["comment"] = webEngine.Group("v1/api/comment")

	AppRouterGroups["profile"] = webEngine.Group("v1/api/profile")
	// 服务器统计和管理相关的路由组
	AppRouterGroups["server"] = webEngine.Group("v1/api/server")
	AppRouterGroups["health"] = webEngine.Group("v1/api/health")
	AppRouterGroups["stats"] = webEngine.Group("v1/api/stats")
}

func getRouterGroupNames() []string {
	names := make([]string, 0, len(AppRouterGroups))
	for name := range AppRouterGroups {
		names = append(names, name)
	}
	return names
}

func InitCors(c *config.Config) {
	// 从配置文件读取CORS origins，如果为空则使用默认值
	corsOrigins := c.Server.CorsOrigins
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"https://lionchat.online", "https://lionchat.online/"}
	}

	webEngine.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

func InitMiddleware() {
	// webEngine.Use(middlewares.RequestID())
	AppRouterGroups["friend"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse).Use()
	AppRouterGroups["group"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["profile"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["moment"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["comment"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
	AppRouterGroups["message"].Use(middlewares.JwtMiddleware()).Use(middlewares.JwtParse)
}

func InitRouter() {
	// *** 用户相关的路由
	AppRouterGroups["user"].POST("/login", api.Login)
	AppRouterGroups["user"].POST("/register", api.Register)
	AppRouterGroups["user"].POST("/logout", api.Logout)
	AppRouterGroups["profile"].GET("/profileInfo", api.GetUserInfor)
	AppRouterGroups["profile"].POST("/profileInfo", api.UpdateUserInfor)

	// ***  socket相关的路由
	AppRouterGroups["webSocket"].GET("/connect", api.WebSocketConnect)

	//*** 好友相关的路由
	AppRouterGroups["friend"].GET("/search", api.SearchClient)
	AppRouterGroups["friend"].POST("/friends", api.AddFriend)
	AppRouterGroups["friend"].POST("/friendResponse", api.HandleFriendResponse)
	AppRouterGroups["friend"].GET("/friendList", api.GetFriendList)

	//*** 消息相关的路由
	AppRouterGroups["message"].GET("/getUnreadMessage", api.GetUnreadMessage)
	AppRouterGroups["message"].POST("/markAsRead", api.MarkMessageAsRead)

	//*** 群组相关的路由
	AppRouterGroups["group"].POST("", api.CreateGroup)
	AppRouterGroups["group"].POST("/group-memberships", api.JoinGroup)
	AppRouterGroups["group"].DELETE("/group-memberships", api.LeaveGroup)
	AppRouterGroups["group"].GET("/group-list", api.GetGroupList)
	AppRouterGroups["group"].GET("/group-members-List", api.GetGroupMembersList)
	AppRouterGroups["group"].PUT("/changeGroupInfo", api.ChangeGroupInfo)

	// *** 监控相关的路由
	AppRouterGroups["monitor"].GET("/clients", api.GetClients)
	AppRouterGroups["monitor"].GET("/serverInfo", api.GetServerInfor)

	AppRouterGroups["moment"].POST("", api.CreateMoment)
	AppRouterGroups["moment"].GET("/moment-list", api.ListMoment)

	AppRouterGroups["comment"].POST("/create", api.CreateComment)
	AppRouterGroups["comment"].POST("/like", api.LikeComment)
	AppRouterGroups["comment"].GET("/list", api.GetCommentList)
}
