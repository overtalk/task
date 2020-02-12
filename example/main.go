package main

import (
	"errors"
	"fmt"
	"log"
	"task"
	"time"
)

type myTask struct {
	name  string
	times int
}

func (myTask *myTask) Execute() error {
	fmt.Printf("*Execute* [%s], times = [%d]\n", myTask.name, myTask.times)

	if myTask.times == 0 {
		myTask.times++
		return errors.New("fail")
	}

	return nil
}

func (myTask *myTask) Success(result *task.Result) {
	fmt.Printf("*success* [%s]! times = [%d]\n", myTask.name, myTask.times)
	//fmt.Println("result.IsSuccessful() = ", result.IsSuccessful())
	//fmt.Println("result.GetError() = ", result.GetError())
	//fmt.Println("result.GetTimeStampMs() = ", result.GetTimeStampMs())
	//fmt.Println("result.GetReservedAttempts() = ", result.GetReservedAttempts())
}

func (myTask *myTask) Fail(result *task.Result) {
	fmt.Printf("*fail* [%s]! times = [%d]\n", myTask.name, myTask.times)
	//fmt.Println("result.IsSuccessful() = ", result.IsSuccessful())
	//fmt.Println("result.GetError() = ", result.GetError())
	//fmt.Println("result.GetTimeStampMs() = ", result.GetTimeStampMs())
	//fmt.Println("result.GetReservedAttempts() = ", result.GetReservedAttempts())
}

func main() {
	c := task.GetDefaultConfig()
	taskPool, err := task.NewTaskPool(c)
	if err != nil {
		log.Fatal(err)
	}
	taskPool.Start()
	defer taskPool.SafeClose()

	for i := 0; i < 10; i++ {
		t := &myTask{
			name:  fmt.Sprintf("task - %d", i),
			times: 0,
		}
		taskPool.PushTask(t)
	}

	time.Sleep(3 * time.Second)
}
