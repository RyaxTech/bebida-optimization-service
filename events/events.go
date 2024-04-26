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

func NewPendingPod(nbCores int, durationInSec int) PendingPod {
	return PendingPod{
		NbCores:       nbCores,
		RequestedTime: time.Duration(durationInSec) * time.Second,
	}
}

type PodCompleted struct {
	PodId        string
	NbCores      int
	TimeCritical bool
}
