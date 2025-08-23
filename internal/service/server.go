package service

import (
	"cchat/pkg/cgoroutinue"
)

type Server struct {
	WorkerHouse *WorkerHouse // 工作者房子
}

var ServerInstance = &Server{
	WorkerHouse: nil,
}

func (s *Server) Run() {
	// 初始化工作者房子
	// 设置默认房子里的工作者数量
	// 后续由客户端建立连接，选择工作者手中持有最少的任务的工作者进行分配
	s.WorkerHouse = InitWorkerHouse(10)
	// 启动工作者
	for _, worker := range s.WorkerHouse.Workers {
		cgoroutinue.GoroutinePool.Submit(func() {
			worker.Do()
		})
	}
}
