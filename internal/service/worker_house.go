package service

import (
	"cchat/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

// 工作者住的房子
type WorkerHouse struct {
	Workers []*Worker
}

// 该怎么从房子里取出一个工作者呢？
// 取出一个工作者，需要根据工作者手中的任务数量，取出任务数量最少的那个工作者
func (h *WorkerHouse) GetWorker() *Worker {
	if len(h.Workers) == 0 {
		return nil
	}
	worker := h.Workers[0]
	for _, w := range h.Workers {
		if w.TaskCount > worker.TaskCount {
			worker = w
		}
	}
	worker.TaskCount--
	logger.Info("取出一个工作者", zap.Int("workerID", worker.ID), zap.Int("taskCount", worker.TaskCount))

	return worker
}

func (h *WorkerHouse) GetWorkerByID(id int) *Worker {
	for _, w := range h.Workers {
		if w.ID == id {
			return w
		}
	}
	return nil
}

// 往房子里增加一个工作者
func (h *WorkerHouse) AddWorker(worker *Worker) {
	h.Workers = append(h.Workers, worker)
}

// 从房子里移除一个工作者
func (h *WorkerHouse) RemoveWorker(worker *Worker) {
	for i, w := range h.Workers {
		if w.ID == worker.ID {
			h.Workers = append(h.Workers[:i], h.Workers[i+1:]...)
			break
		}
	}
}

func InitWorkerHouse(workerNum int) *WorkerHouse {
	workerHouse := &WorkerHouse{
		Workers: make([]*Worker, workerNum),
	}
	for i := 0; i < workerNum; i++ {
		worker := &Worker{
			ID:              i,
			Clients:         sync.Map{},
			Register:        make(chan *Client, 100),
			Unregister:      make(chan *Client, 100),
			Broadcast:       make(chan []byte, 100),
			mutex:           sync.RWMutex{},
			FragmentManager: NewFragmentManager(),
			TaskCount:       10,
			WorkerHouse:     workerHouse,
		}
		workerHouse.Workers[i] = worker
	}
	return workerHouse
}
