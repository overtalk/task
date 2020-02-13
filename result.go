package task

type Attempt struct {
	Success     bool
	Error       error
	TimeStampMs int64
}

func newAttempt(success bool, err error) *Attempt {
	return &Attempt{
		Success:     success,
		Error:       err,
		TimeStampMs: getTimeMs(),
	}
}

type Result struct {
	task        Task
	successful  bool
	attemptList []*Attempt
}

func (result *Result) GetTask() Task {
	return result.task
}

func (result *Result) IsSuccessful() bool {
	return result.successful
}

func (result *Result) GetReservedAttempts() []*Attempt {
	return result.attemptList
}

func (result *Result) GetError() error {
	if len(result.attemptList) == 0 {
		return nil
	}
	cursor := len(result.attemptList) - 1
	return result.attemptList[cursor].Error
}

func (result *Result) GetTimeStampMs() int64 {
	if len(result.attemptList) == 0 {
		return 0
	}
	cursor := len(result.attemptList) - 1
	return result.attemptList[cursor].TimeStampMs
}

func initResult(t Task) *Result {
	return &Result{
		attemptList: []*Attempt{},
		successful:  false,
		task:        t,
	}
}
