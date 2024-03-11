package main

import (
	"os"
	"strconv"

	connectors "github.com/RyaxTech/bebida-shaker/connectors"
	"github.com/RyaxTech/bebida-shaker/events"
	"github.com/apex/log"
)

type HPCSchedulerType string

type Parameters struct {
	maxPendingPunchJob int
	HPCSchedulerType   string
}

var params Parameters
var podIdToJobIdMap = make(map[string]string)

func schedule(event interface{}, hpcScheduler connectors.HPCConnector) {
	switch event := event.(type) {
	case events.PendingPod:
		log.Infof("Handling new pending pod: %v+\n", event)
		if event.Deadline.IsZero() && len(podIdToJobIdMap) >= params.maxPendingPunchJob {
			log.Warnf("Do not create Punch job because we reach the max number of punch job on this cluster: %d)", params.maxPendingPunchJob)
			return
		}
		jobId, err := hpcScheduler.Punch(int(event.NbCores), event.RequestedTime, event.Deadline)
		podIdToJobIdMap[event.PodId] = jobId
		if err != nil {
			log.Errorf("Unable to allocate resources %s", err)
		}
	case events.PodCompleted:
		log.Infof("Handling pod completed: %v+\n", event)
		hpcScheduler.QuitPunch(podIdToJobIdMap[event.PodId])
	default:
		log.Fatalf("Unknown event %v+\n", event)
		panic(-1)
	}

}

func run() {
	event_channel := make(chan interface{})

	var HPCScheduler connectors.HPCConnector
	switch params.HPCSchedulerType {
	case "OAR":
		HPCScheduler = connectors.OAR{}
	case "SLURM":
		HPCScheduler = connectors.SLURM{}
	}

	go connectors.WatchQueues(event_channel)
	for {
		event := <-event_channel
		schedule(event, HPCScheduler)
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
	log.Info("Starting Bebida Shaker")
	params = Parameters{
		maxPendingPunchJob: getIntEnv("BEBIDA_MAX_PENDING_PUNCH_JOB", 2),
		HPCSchedulerType:   getStrEnv("BEBIDA_HPC_SCHEDULER_TYPE", "OAR"),
	}
	log.Infof("Parameters: %+v\n", params)
	run()
}
