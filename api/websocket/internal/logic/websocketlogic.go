package logic

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"

	"chatLion/api/websocket/internal/svc"
	"chatLion/api/websocket/internal/types"
	jsoncontent "chatLion/data"
	myjwt "chatLion/jwt"
	"chatLion/rpc/group/group"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type WebsocketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	w      http.ResponseWriter
	r      *http.Request
}

func NewWebsocketLogic(ctx context.Context, svcCtx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) *WebsocketLogic {
	return &WebsocketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		w:      w,
		r:      r,
	}
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var connectionPool sync.Map

func (l *WebsocketLogic) Websocket(req *types.WebSocketRequest) (resp *types.WebSocketResponse, err error) {
	// 生成一个websocket连接对象
	websocketConnectionInstance, err := upgrade.Upgrade(l.w, l.r, nil)
	if err != nil {
		return &types.WebSocketResponse{
			Message: "websocket 升级失败",
		}, err
	}
	defer websocketClose(websocketConnectionInstance)

	jwt := l.r.Header.Get("Authorization")
	// 验证jwt
	jwtUnencoder := myjwt.JWTUnencoder([]byte("chatLion"), jwt)
	// 保存连接到连接池
	connectionPool.Store(jwtUnencoder.Email, websocketConnectionInstance)
	// 处理消息循环
	// todo : 是否使用goroutine处理消息循环？
	for {
		currentMessage := new(jsoncontent.MessageContent)
		err := websocketConnectionInstance.ReadJSON(currentMessage)
		// 判断消息的类型：0：单聊， 1：群聊
		if err != nil {
			log.Println("消息json解析出错")
			continue
		}
		if currentMessage.MessageType == 0 {
			websocketSendMessageToOne(currentMessage.To, currentMessage.Content)
		} else if currentMessage.MessageType == 1 {
			// 远程调用获取群成员列表
			groupResp, err := l.svcCtx.GroupRPC.GetMembersByGroupID(l.ctx, &group.GetMembersRequest{
				GroupId: currentMessage.To,
			})
			if len(groupResp.Members) == 0 {
				log.Println("群成员为空")
				continue
			}
			if err != nil {
				log.Println("获取群成员失败")
				continue
			}
			websocketSendMessageToGroup(currentMessage.From, groupResp.Members, currentMessage.Content)
		} else {
			log.Println("无效的消息类型")
		}
	}
}

// 发送消息到指定的群聊
func websocketSendMessageToGroup(from string, to []string, message string) error {
	for _, target := range to {
		// 从连接池中获取连接
		target = strings.TrimSpace(target)
		targetConn, ok := connectionPool.Load(target)
		if !ok || target == from {
			continue
		}
		conn := targetConn.(*websocket.Conn)
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("消息发送失败")
			return err
		}
	}
	return nil
}

// 发送消息到指定的用户
func websocketSendMessageToOne(uuid string, message string) error {
	// 从连接池中获取连接
	targetConn, ok := connectionPool.Load(uuid)
	if !ok {
		return nil
	}
	conn := targetConn.(*websocket.Conn)
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func websocketClose(conn *websocket.Conn) {
	conn.Close()
}
