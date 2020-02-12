package task

import (
	"container/heap"
	"sync"
)

type retryQueue struct {
	retrySlice
	retryQueueShutDownFlag bool
	retryQueueWait         *sync.WaitGroup
}

func newRetryQueue(retryQueueWait *sync.WaitGroup) *retryQueue {
	retrySlice := retrySlice{}
	heap.Init(&retrySlice)

	return &retryQueue{
		retrySlice:             retrySlice,
		retryQueueShutDownFlag: false,
		retryQueueWait:         retryQueueWait,
	}
}

func (retryQueue *retryQueue) sendToRetryQueue(task *task) {
	if task != nil {
		heap.Push(retryQueue, task)
	}
}

func (retryQueue *retryQueue) getRetryTaskList() []*task {
	var taskList []*task

	for retryQueue.Len() > 0 {
		t := heap.Pop(retryQueue)
		if t.(*task).nextRetryMs < getTimeMs() {
			taskList = append(taskList, t.(*task))
		} else {
			heap.Push(retryQueue, t.(*task))
			break
		}
	}

	return taskList
}

type retrySlice []*task

func (retryQueue retrySlice) Len() int {
	return len(retryQueue)
}

func (retryQueue retrySlice) Less(i, j int) bool {
	return retryQueue[i].nextRetryMs < retryQueue[j].nextRetryMs
}
func (retryQueue retrySlice) Swap(i, j int) {
	retryQueue[i], retryQueue[j] = retryQueue[j], retryQueue[i]
}
func (retryQueue *retrySlice) Push(x interface{}) {
	item := x.(*task)
	*retryQueue = append(*retryQueue, item)
}
func (retryQueue *retrySlice) Pop() interface{} {
	old := *retryQueue
	n := len(old)
	item := old[n-1]
	*retryQueue = old[0 : n-1]
	return item
}
