package api

import (
	"cchat/internal/dto"
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

// ✅
func WebSocketConnect(c *gin.Context) {
	access_token := c.Query("token")
	if access_token == "" {
		c.JSON(401, dto.WebSocketResponse{
			Code: 1019,
			Msg:  "token 无效",
		})
		return
	}
	// 解析token
	claims, err := token.ParseToken(access_token)
	if err != nil {
		c.JSON(401, dto.WebSocketResponse{
			Code: 1019,
			Msg:  "token 无效",
		})
		return
	}
	// 从token中获取用户id
	uuid := claims.UserUUID
	if uuid == "" {
		c.JSON(401, dto.WebSocketResponse{
			Code: 1019,
			Msg:  "token 无效",
		})
		return
	}

	// 连接websocket http -> websocket
	ws, err := UpdateGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.WebSocketResponse{
			Code: 1025,
			Msg:  "websocket连接失败",
		})
		return
	}
	wserr := service.RegisterClient(ws, uuid)
	if wserr != nil {
		c.JSON(http.StatusInternalServerError, dto.WebSocketResponse{
			Code: wserr.Code,
			Msg:  wserr.Msg,
		})
		return
	}
	c.JSON(http.StatusOK, dto.WebSocketResponse{
		Code: 0,
		Msg:  "websocket连接成功",
	})
}
