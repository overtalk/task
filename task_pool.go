package task

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	IllegalStateExceptionErr = errors.New("IllegalStateException")
	TimeoutExceptionErr      = errors.New("TimeoutException")
)

type TaskPool struct {
	lock                 sync.RWMutex
	taskPoolShutDownFlag bool
	factory              *taskFactory
	queue                []*task
	retryQueue           *retryQueue
	ioWorker             *ioWorker
	maxTaskNum           int
	MaxBlockSec          int
	// wg
	retryQueueWait *sync.WaitGroup
	taskPoolWait   *sync.WaitGroup
	ioWorkerWait   *sync.WaitGroup
}

func NewTaskPool(c *Config) (*TaskPool, error) {
	taskPool := &TaskPool{
		lock:                 sync.RWMutex{},
		taskPoolShutDownFlag: false,
		maxTaskNum:           c.MaxTaskNum,
		queue:                []*task{},
		retryQueueWait:       &sync.WaitGroup{},
		taskPoolWait:         &sync.WaitGroup{},
		ioWorkerWait:         &sync.WaitGroup{},
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

// Limited closing transfer parameter nil, safe closing transfer timeout time, timeout Ms parameter in milliseconds
func (taskPool *TaskPool) Close(timeoutMs int64) error {
	startCloseTime := time.Now().Unix()
	taskPool.retryQueue.retryQueueShutDownFlag = true
	taskPool.retryQueueWait.Wait()
	taskPool.taskPoolShutDownFlag = true
	for {
		taskCount := atomic.LoadInt64(&taskPool.ioWorker.taskCount)
		if taskCount != 0 && time.Now().Unix()-startCloseTime <= timeoutMs {
			time.Sleep(time.Second)
		} else if taskCount == 0 && len(taskPool.queue) == 0 {
			return nil
		} else if time.Now().Unix()-startCloseTime > timeoutMs {
			return IllegalStateExceptionErr
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (taskPool *TaskPool) SafeClose() {
	taskPool.retryQueue.retryQueueShutDownFlag = true
	taskPool.retryQueueWait.Wait()
	taskPool.taskPoolShutDownFlag = true
	taskPool.taskPoolWait.Wait()
	taskPool.ioWorkerWait.Wait()
}

func (taskPool *TaskPool) PushTask(task Task) error {
	if err := taskPool.waitTime(); err != nil {
		return err
	}

	taskPool.pushTask(taskPool.factory.produce(task))
	return nil
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

func (taskPool *TaskPool) taskNum() int {
	return 0
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
			time.Sleep(100 * time.Millisecond)
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

func (taskPool *TaskPool) waitTime() error {
	if taskPool.MaxBlockSec > 0 {
		for i := 0; i < taskPool.MaxBlockSec; i++ {
			if taskPool.taskNum() > taskPool.maxTaskNum {
				time.Sleep(time.Second)
			} else {
				return nil
			}
		}
		return TimeoutExceptionErr
	} else if taskPool.MaxBlockSec == 0 {
		if taskPool.taskNum() > taskPool.maxTaskNum {
			return TimeoutExceptionErr
		}
	} else if taskPool.MaxBlockSec < 0 {
		for {
			if taskPool.taskNum() > taskPool.maxTaskNum {
				time.Sleep(time.Second)
			} else {
				return nil
			}
		}
	}
	return nil
}
