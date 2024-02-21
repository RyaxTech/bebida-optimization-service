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
	threshold         int
	pendingJobs       int
	maxPendingJob     int
	stepTimeInSeconds int
	HPCSchedulerType  string
}

var params Parameters

func schedule(event interface{}, hpcScheduler connectors.HPCConnector) {
	switch event.(type) {
	case events.NewPendingPod:
		log.Infof("Handling new pending pod: %v+\n", event)
		pod := event.(events.NewPendingPod)
		_, err := hpcScheduler.Punch(int(pod.NbCores), int(pod.Requested_time.Seconds()))
		if err != nil {
			log.Errorf("Unable to allocate resources %s", err)
		}
	case events.PodCompleted:
		log.Infof("Handling pod completed: %v+\n", event)

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
		threshold:         getIntEnv("BEBIDA_NB_PENDING_JOB_THRESHOLD", 1),
		pendingJobs:       0,
		maxPendingJob:     getIntEnv("BEBIDA_MAX_PENDING_PUNCH_JOB", 1),
		HPCSchedulerType:  getStrEnv("BEBIDA_HPC_SCHEDULER_TYPE", "OAR"),
		stepTimeInSeconds: getIntEnv("BEBIDA_STEP_IN_SECONDS", 3),
	}
	log.Infof("Parameters: %+v\n", params)
	run()
}
