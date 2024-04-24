package events

import "time"

// events
type PendingPod struct {
	PodId         string
	NbCores       int
	RequestedTime time.Duration
	Deadline      time.Time
	TimeCritical  bool
}

func NewPendingPod() PendingPod {
	return PendingPod{
		NbCores:       1,
		RequestedTime: 900 * time.Second,
	}
}

type PodCompleted struct {
	PodId          string
	NbCore         int
	CompletionTime time.Duration
	TimeCritical   bool
}
