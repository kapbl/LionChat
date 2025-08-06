package api

import (
	"cchat/internal/service"
	"cchat/pkg/token"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var UpdateGrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketConnect(c *gin.Context) {
	// 从请求参数中获取jwt
	inputToken := c.Query("token")
	// 解析jwt
	claims, err := token.ParseToken(inputToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 1,
			"msg":  "token 无效",
		})
		return
	}
	uuid := claims.UserUUID
	// 连接websocket http -> websocket
	ws, err := UpdateGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "连接失败",
		})
		return
	}
	// 在服务器中注册成为一个客户端
	client := service.NewClient(ws, uuid)
	// 选择作坊里手中任务最少的那个工作者进行分配任务
	worker := service.ServerInstance.WorkerHouse.GetWorker()

	if worker == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 1,
			"msg":  "连接失败",
		})
		return
	}
	worker.Register <- client

}
