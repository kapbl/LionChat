package cgoroutinue

// 任务
type Task struct {
	f func() error
}

func CreateTask(f func() error) *Task {
	return &Task{
		f: f,
	}
}

func (t *Task) Run() error {
	return t.f()
}


// 协程池
type GPool struct{
	TaskChannel chan *Task
	WorkerNum int
	JobsChannel chan *Task
}

func (g *GPool) worker(workerID int) {
	for task := range g.JobsChannel{
		task.Run()
	}
}

// 创建一个协程池
func (g *GPool) CreatePool() {
	for i := 0; i < g.WorkerNum; i++ {
		go g.worker(i)
	}
}


