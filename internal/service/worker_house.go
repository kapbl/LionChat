package service

import (
	"cchat/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

// 工作者住的房子
type WorkerHouse struct {
	Workers    []*Worker
	mutex      sync.RWMutex // 保护Workers切片的读写锁
	maxWorkers int          // 最大Worker数量限制
	threshold  int          // 触发扩容的任务数量阈值
}

// 该怎么从房子里取出一个工作者呢？
// 取出一个工作者，需要根据工作者手中的任务数量，取出任务数量最少的那个工作者
// 如果所有工作者都满载，则自动扩容一个新的工作者
func (h *WorkerHouse) GetWorker() *Worker {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.Workers) == 0 {
		return nil
	}

	// 检查是否需要扩容
	if h.shouldExpand() {
		h.expandWorkerPool()
	}

	worker := h.Workers[0]
	for _, w := range h.Workers {
		if w.TaskCount < worker.TaskCount {
			worker = w
		}
	}
	worker.TaskCount++
	logger.Info("取出一个工作者", zap.Int("workerID", worker.ID), zap.Int("taskCount", worker.TaskCount))

	return worker
}

func (h *WorkerHouse) GetWorkerByID(id int) *Worker {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, w := range h.Workers {
		if w.ID == id {
			return w
		}
	}
	return nil
}

// 往房子里增加一个工作者
func (h *WorkerHouse) AddWorker(worker *Worker) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.Workers = append(h.Workers, worker)
}

// 从房子里移除一个工作者
func (h *WorkerHouse) RemoveWorker(worker *Worker) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for i, w := range h.Workers {
		if w.ID == worker.ID {
			h.Workers = append(h.Workers[:i], h.Workers[i+1:]...)
			break
		}
	}
}

// shouldExpand 检查是否需要扩容
// 当所有Worker的TaskCount都超过阈值时，触发扩容
func (h *WorkerHouse) shouldExpand() bool {
	// 如果已达到最大Worker数量，不再扩容
	if len(h.Workers) >= h.maxWorkers {
		return false
	}

	// 检查所有Worker是否都满载
	for _, worker := range h.Workers {
		if worker.TaskCount < h.threshold {
			return false
		}
	}

	logger.Info("所有Worker都满载，准备扩容",
		zap.Int("currentWorkers", len(h.Workers)),
		zap.Int("threshold", h.threshold))
	return true
}

// expandWorkerPool 扩容Worker池
func (h *WorkerHouse) expandWorkerPool() {
	newWorkerID := len(h.Workers)
	worker := &Worker{
		ID:              newWorkerID,
		Clients:         sync.Map{},
		Register:        make(chan *Client, 100),
		Unregister:      make(chan *Client, 100),
		Broadcast:       make(chan []byte, 100),
		mutex:           sync.RWMutex{},
		FragmentManager: NewFragmentManager(),
		TaskCount:       0, // 新Worker初始任务数为0
		WorkerHouse:     h,
		MessageQueue:    make(chan *MessageTask, 100),
	}

	h.Workers = append(h.Workers, worker)

	// 启动新Worker
	go worker.Do()

	logger.Info("成功扩容新Worker",
		zap.Int("newWorkerID", newWorkerID),
		zap.Int("totalWorkers", len(h.Workers)))
}

// GetWorkerStats 获取Worker统计信息
func (h *WorkerHouse) GetWorkerStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["totalWorkers"] = len(h.Workers)
	stats["maxWorkers"] = h.maxWorkers
	stats["threshold"] = h.threshold

	workerStats := make([]map[string]interface{}, 0, len(h.Workers))
	for _, worker := range h.Workers {
		workerStat := map[string]interface{}{
			"id":        worker.ID,
			"taskCount": worker.TaskCount,
			"clientCount": func() int {
				count := 0
				worker.Clients.Range(func(_, _ interface{}) bool {
					count++
					return true
				})
				return count
			}(),
		}
		workerStats = append(workerStats, workerStat)
	}
	stats["workers"] = workerStats

	return stats
}

func InitWorkerHouse(workerNum int) *WorkerHouse {
	return InitWorkerHouseWithConfig(workerNum, workerNum*2, 150)
}

// InitWorkerHouseWithConfig 使用配置初始化WorkerHouse
func InitWorkerHouseWithConfig(initialWorkers, maxWorkers, threshold int) *WorkerHouse {
	workerHouse := &WorkerHouse{
		Workers:    make([]*Worker, 0, initialWorkers),
		maxWorkers: maxWorkers,
		threshold:  threshold,
	}

	for i := 0; i < initialWorkers; i++ {
		worker := &Worker{
			ID:              i,
			Clients:         sync.Map{},
			Register:        make(chan *Client, 100),
			Unregister:      make(chan *Client, 100),
			Broadcast:       make(chan []byte, 100),
			mutex:           sync.RWMutex{},
			FragmentManager: NewFragmentManager(),
			TaskCount:       0, // 初始任务数量为0
			WorkerHouse:     workerHouse,
			MessageQueue:    make(chan *MessageTask, 100),
		}
		workerHouse.Workers = append(workerHouse.Workers, worker)
	}

	logger.Info("WorkerHouse初始化完成",
		zap.Int("initialWorkers", initialWorkers),
		zap.Int("maxWorkers", maxWorkers),
		zap.Int("threshold", threshold))

	return workerHouse
}
