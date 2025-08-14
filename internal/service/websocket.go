package service

import (
	"cchat/pkg/cerror"

	"github.com/gorilla/websocket"
)

// ✅
func RegisterClient(ws *websocket.Conn, uuid string) *cerror.CodeError {
	// 在服务器中注册成为一个客户端
	client := NewClient(ws, uuid)
	// 选择作坊里手中任务最少的那个工作者进行分配任务
	worker := ServerInstance.WorkerHouse.GetWorker()
	if worker == nil {
		return cerror.NewCodeError(1025, "连接失败")
	}
	worker.Register <- client
	return nil
}
