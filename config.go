package task

type Config struct {
	// for factory
	MaxReservedAttempts int   // 最大剩余尝试次数
	MaxRetryTimes       int   // 最大尝试次数
	BaseRetryBackOffMs  int64 // 首次重试的退避时间
	MaxRetryBackOffMs   int64 // 重试的最大退避时间，默认为 50 秒
	// for worker
	MaxIoWorkerNum int // 最多worker数量（协程数量）
	MaxTaskNum     int // 最多任务数量
	MaxBlockSec    int // 最大阻塞时间
}

func GetDefaultConfig() *Config {
	return &Config{
		MaxReservedAttempts: 11,
		MaxRetryTimes:       10,
		BaseRetryBackOffMs:  100,
		MaxRetryBackOffMs:   50 * 1000,
		MaxIoWorkerNum:      50,
		MaxTaskNum:          1000,
		MaxBlockSec:         60,
	}
}
