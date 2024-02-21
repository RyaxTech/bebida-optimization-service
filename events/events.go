package events

import "time"

// events
type NewPendingPod struct {
	PodId          string
	NbCores         int
	Requested_time time.Duration
	Deadline       time.Time
	TimeCritical bool
}
type PodCompleted struct {
	PodId           string
	NbCore          int
	Completion_time time.Duration
	TimeCritical bool
}
