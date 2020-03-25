package main

import (
	"errors"
	"fmt"
	"github.com/overtalk/task"
	"log"
	"time"
)

type a interface {
	a()
}

type B interface {
	a
	b()
}

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

func (myTask *myTask) CallBack(result *task.Result) {
	if result.IsSuccessful() {
		fmt.Printf("*success* [%s]! times = [%d]\n", myTask.name, myTask.times)
	} else {
		fmt.Printf("*fail* [%s]! times = [%d]\n", myTask.name, myTask.times)
	}
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
