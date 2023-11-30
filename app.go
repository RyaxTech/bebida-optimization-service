package main

import (
	"os"
	"time"

	connectors "github.com/RyaxTech/bebida-shaker/connectors"
	"github.com/apex/log"
)

type HPCSchedulerType string

type Parameters struct {
	threshold        int
	pendingJobs      int
	maxPendingJob    int
	HPCSchedulerType string
}

var params Parameters

// Simulate a function that takes 1s to complete.
func run() string {
	log.Info("Check for the Queue state")
	queueSize, err := connectors.GetQueueSize()
	if err != nil {
		log.Errorf("Unable to get size the queue %s", err)
	}
	nbRunningApp, err := connectors.GetNbRunningApp()
	if err != nil {
		log.Errorf("Unable to get number of running app %s", err)
	}

	var HPCScheduler connectors.HPConnector
	switch params.HPCSchedulerType {
	case "OAR":
		HPCScheduler = connectors.OAR{}
	case "SLURM":
		HPCScheduler = connectors.SLURM{}
	}

	var jobID string

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
		HPCScheduler.QuitPunch(jobID)
	}
	return ""
}

func RunForever() {
	for {
		go run()
		time.Sleep(1 * time.Second)
	}
}

func main() {
	log.Info("Starting Bebida Shacker")
	params = Parameters{threshold: 1, pendingJobs: 0, maxPendingJob: 1, HPCSchedulerType: os.Getenv("BEBIDA_HPC_SCHEDULER_TYPE")}
	log.Infof("Parameters: %+v\n", params)
	RunForever()
}
