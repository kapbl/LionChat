package service

import (
	"cchat/pkg/cgoroutinue"
	"sync"
)

type Server struct {
	Clients         sync.Map         // 存储所有客户端连接
	mutex           *sync.Mutex      // 保护clients的并发访问
	Broadcast       chan []byte      // 广播通道
	Register        chan *Client     // 注册消息通道
	Ungister        chan *Client     // 注销消息通道
	FragmentManager *FragmentManager // 消息分片管理器
	WorkerHouse     *WorkerHouse     // 工作者房子
}

var ServerInstance = &Server{
	Clients:         sync.Map{},
	mutex:           &sync.Mutex{},
	Broadcast:       make(chan []byte, 1000),
	Register:        make(chan *Client, 100),
	Ungister:        make(chan *Client, 100),
	FragmentManager: NewFragmentManager(),
	WorkerHouse:     nil,
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
