package task

import (
	"math"
	"sync"
	"sync/atomic"
)

type ioWorker struct {
	ioWorkerWait *sync.WaitGroup
	retryQueue   *retryQueue
	taskCount    int64         // current task count
	maxIoWorker  chan struct{} // to control goroutine num
}

func newIoWorker(maxIoWorkerNum int, retryQueue *retryQueue, ioWorkerWait *sync.WaitGroup) (*ioWorker, error) {
	return &ioWorker{
		ioWorkerWait: ioWorkerWait,
		retryQueue:   retryQueue,
		taskCount:    0,
		maxIoWorker:  make(chan struct{}, maxIoWorkerNum),
	}, nil
}

func (ioWorker *ioWorker) DoTask(t *task) {
	defer ioWorker.finishTask()
	atomic.AddInt64(&ioWorker.taskCount, 1)

	if err := t.task.Execute(); err == nil {
		if t.attemptCount < t.maxReservedAttempts {
			attempt := newAttempt(true, nil)
			t.result.attemptList = append(t.result.attemptList, attempt)
		}
		t.result.successful = true
		t.task.Success(t.result)
	} else {
		// if the retry queue is already closed
		if ioWorker.retryQueue.retryQueueShutDownFlag {
			t.task.Fail(t.result)
			return
		}

		if t.attemptCount < t.maxRetryTimes {
			if t.attemptCount < t.maxReservedAttempts {
				attempt := newAttempt(false, err)
				t.result.attemptList = append(t.result.attemptList, attempt)
			}
			t.result.successful = false
			t.attemptCount += 1
			retryWaitTime := t.baseRetryBackOffMs * int64(math.Pow(2, float64(t.attemptCount)-1))
			if retryWaitTime < t.maxRetryIntervalInMs {
				t.nextRetryMs = getTimeMs() + retryWaitTime
			} else {
				t.nextRetryMs = getTimeMs() + t.maxRetryIntervalInMs
			}
			ioWorker.retryQueue.sendToRetryQueue(t)
		} else {
			t.task.Fail(t.result)
		}
	}
}

func (ioWorker *ioWorker) finishTask() {
	ioWorker.ioWorkerWait.Done()
	atomic.AddInt64(&ioWorker.taskCount, -1)
	<-ioWorker.maxIoWorker
}

func (ioWorker *ioWorker) CurrentTaskNum() int64 {
	return atomic.LoadInt64(&ioWorker.taskCount)
}
