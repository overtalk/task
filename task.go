package task

type Task interface {
	Execute() error
	Success(result *Result)
	Fail(result *Result)
}

// task is used in the project
type task struct {
	attemptCount         int     // 尝试次数
	maxReservedAttempts  int     // 最大剩余尝试次数
	maxRetryTimes        int     // 最大尝试次数
	baseRetryBackOffMs   int64   // 首次重试的退避时间
	maxRetryIntervalInMs int64   // 重试的最大退避时间，默认为 50 秒
	createTimeMs         int64   // 创建的时间
	nextRetryMs          int64   // 下次重试的时间
	result               *Result // 发送结果
	task                 Task
}

// *************************
// task factory
type taskFactory struct {
	maxReservedAttempts  int   // 最大剩余尝试次数
	maxRetryTimes        int   // 最大尝试次数
	baseRetryBackOffMs   int64 // 首次重试的退避时间
	maxRetryIntervalInMs int64 // 重试的最大退避时间，默认为 50 秒
}

func newTaskFactory(c *Config) *taskFactory {
	return &taskFactory{
		maxReservedAttempts:  c.MaxReservedAttempts,
		maxRetryTimes:        c.MaxRetryTimes,
		baseRetryBackOffMs:   c.BaseRetryBackOffMs,
		maxRetryIntervalInMs: c.MaxRetryBackOffMs,
	}
}

func (taskFactory *taskFactory) produce(t Task) *task {
	return &task{
		attemptCount:         0,
		maxReservedAttempts:  taskFactory.maxReservedAttempts,
		maxRetryTimes:        taskFactory.maxRetryTimes,
		baseRetryBackOffMs:   taskFactory.baseRetryBackOffMs,
		maxRetryIntervalInMs: taskFactory.maxRetryIntervalInMs,
		createTimeMs:         getTimeMs(),
		nextRetryMs:          0,
		result:               initResult(t),
		task:                 t,
	}
}
