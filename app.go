package main

import (
	"os"
	"strconv"
	"time"

	connectors "github.com/RyaxTech/bebida-shaker/connectors"
	"github.com/apex/log"
)

type HPCSchedulerType string

type Parameters struct {
	threshold         int
	pendingJobs       int
	maxPendingJob     int
	stepTimeInSeconds int
	HPCSchedulerType  string
}

var params Parameters

// Simulate a function that takes 1s to complete.
func run() string {
	log.Info("Check for the Queue state")
	queueSize, timeCriticalQueueSize, deadlineAwareQueue, err := connectors.GetQueueSize()
	if err != nil {
		log.Errorf("Unable to get size the queue %s", err)
	}
	nbRunningApp, err := connectors.GetNbRunningApp()
	if err != nil {
		log.Errorf("Unable to get number of running app %s", err)
	}

	var HPCScheduler connectors.HPCConnector
	switch params.HPCSchedulerType {
	case "OAR":
		HPCScheduler = connectors.OAR{}
	case "SLURM":
		HPCScheduler = connectors.SLURM{}
	}

	if timeCriticalQueueSize > 0 {
		HPCScheduler.Refill(timeCriticalQueueSize)
	} else {
		HPCScheduler.Refill(-1)
	}

	for _, job := range deadlineAwareQueue {
		log.Debugf("Pending Deadline aware job %+v\n", job)
		jobID, err := HPCScheduler.Punch(int(job.NbCPU), job.Duration_in_seconds)
		if err != nil {
			log.Errorf("Unable to allocate resources %s", err)
		}
		// FIXME: might return multiple job id...
		return jobID
	}

	log.Infof("Queue size found: %d", queueSize)
	log.Infof("Nb running app found: %d", nbRunningApp)
	if queueSize > params.threshold && params.pendingJobs < params.maxPendingJob {
		log.Info("Hummmm... a Ti'Punch ^^")
		params.pendingJobs += 1
		jobID, err := HPCScheduler.Punch(1, 900)
		if err != nil {
			log.Errorf("Unable to allocate resources %s", err)
		}
		params.pendingJobs -= 1
		return jobID
	} else if queueSize == 0 && nbRunningApp == 0 {
		HPCScheduler.QuitAllPunch()
	}
	return ""
}

func RunForever(step time.Duration) {
	punchJobIds := []string{}
	for {
		punchJobId := run()
		if punchJobId != "" {
			punchJobIds = append(punchJobIds, punchJobId)
		}
		time.Sleep(step * time.Second)
	}
}

func getIntEnv(envName string, defaultValue int) int {
	val, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	} else {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			log.Warnf("Unable to parse '%s' environment variable with value '%s': %s", envName, val, err)
			return defaultValue
		}
		return intVal
	}
}

func getStrEnv(envName string, defaultValue string) string {
	val, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	} else {
		return val
	}
}

func main() {
	log.Info("Starting Bebida Shacker")
	params = Parameters{
		threshold:         getIntEnv("BEBIDA_NB_PENDING_JOB_THRESHOLD", 1),
		pendingJobs:       0,
		maxPendingJob:     getIntEnv("BEBIDA_MAX_PENDING_PUNCH_JOB", 1),
		HPCSchedulerType:  getStrEnv("BEBIDA_HPC_SCHEDULER_TYPE", "OAR"),
		stepTimeInSeconds: getIntEnv("BEBIDA_STEP_IN_SECONDS", 3),
	}
	log.Infof("Parameters: %+v\n", params)
	RunForever(time.Duration(params.stepTimeInSeconds))
}
