package task

import "time"

func getTimeMs() int64 {
	return time.Now().UnixNano() / 1000 / 1000
}
