package task

import (
	"sync"
	"time"
)

type TaskPool struct {
	lock                 sync.RWMutex
	taskPoolShutDownFlag bool
	lingerMs             int64
	factory              *taskFactory
	queue                []*task
	retryQueue           *retryQueue
	ioWorker             *ioWorker
	// wg
	retryQueueWait *sync.WaitGroup
	taskPoolWait   *sync.WaitGroup
	ioWorkerWait   *sync.WaitGroup
}

func NewTaskPool(c *Config) (*TaskPool, error) {
	taskPool := &TaskPool{
		lock:                 sync.RWMutex{},
		taskPoolShutDownFlag: false,
		queue:                []*task{},
		retryQueueWait:       &sync.WaitGroup{},
		taskPoolWait:         &sync.WaitGroup{},
		ioWorkerWait:         &sync.WaitGroup{},
		lingerMs:             c.LingerMs,
	}

	factory := newTaskFactory(c)
	retryQueue := newRetryQueue(taskPool.retryQueueWait)
	worker, err := newIoWorker(c.MaxIoWorkerNum, retryQueue, taskPool.ioWorkerWait)
	if err != nil {
		return nil, err
	}

	taskPool.factory = factory
	taskPool.retryQueue = retryQueue
	taskPool.ioWorker = worker

	return taskPool, nil
}

func (taskPool *TaskPool) Start() {
	taskPool.retryQueueWait.Add(1)
	taskPool.taskPoolWait.Add(1)
	go taskPool.start()
}

func (taskPool *TaskPool) SafeClose() {
	taskPool.retryQueue.retryQueueShutDownFlag = true
	taskPool.retryQueueWait.Wait()
	taskPool.taskPoolShutDownFlag = true
	taskPool.taskPoolWait.Wait()
	taskPool.ioWorkerWait.Wait()
}

func (taskPool *TaskPool) PushTask(task Task) {
	taskPool.pushTask(taskPool.factory.produce(task))
}

func (taskPool *TaskPool) pushTask(task *task) {
	defer taskPool.lock.Unlock()
	taskPool.lock.Lock()
	taskPool.queue = append(taskPool.queue, task)
}

func (taskPool *TaskPool) popTask() *task {
	defer taskPool.lock.Unlock()
	taskPool.lock.Lock()
	task := taskPool.queue[0]
	taskPool.queue = taskPool.queue[1:]
	return task
}

func (taskPool *TaskPool) start() {
	go taskPool.retry()

	for {
		if len(taskPool.queue) > 0 {
			select {
			case taskPool.ioWorker.maxIoWorker <- struct{}{}:
				taskPool.ioWorkerWait.Add(1)
				go taskPool.ioWorker.DoTask(taskPool.popTask())
			}
		} else {
			if !taskPool.taskPoolShutDownFlag {
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}
	}

	taskPool.taskPoolWait.Done()
}

func (taskPool *TaskPool) retry() {
	for !taskPool.retryQueue.retryQueueShutDownFlag {
		retryTaskList := taskPool.retryQueue.getRetryTaskList()
		if retryTaskList == nil {
			// If there is nothing to send in the retry queue, just wait for the minimum time that was given to me last time.
			time.Sleep(time.Duration(taskPool.lingerMs) * time.Millisecond)
		} else {
			count := len(retryTaskList)
			for i := 0; i < count; i++ {
				taskPool.pushTask(retryTaskList[i])
			}
		}
	}

	// send task in retry queue to send again
	retryTaskList := taskPool.retryQueue.getRetryTaskList()
	count := len(retryTaskList)
	for i := 0; i < count; i++ {
		taskPool.pushTask(retryTaskList[i])
	}

	taskPool.retryQueue.retryQueueWait.Done()
}
