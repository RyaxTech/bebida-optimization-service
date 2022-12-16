package main

import (
	"time"

	connectors "github.com/RyaxTech/bebida-optimization-service/connectors"
	"github.com/apex/log"
)

type Parameters struct {
	threshold     int
	pendingJobs   int
	maxPendingJob int
}

var params Parameters

// Simulate a function that takes 1s to complete.
func run() {
	log.Info("Check for the Queue state")
	queueSize, err := connectors.GetQueueSize()
	if err != nil {
		log.Errorf("Unable to get size the queue %s", err)
	}
	nbRunningApp, err := connectors.GetNbRunningApp()
	if err != nil {
		log.Errorf("Unable to get number of running app %s", err)
	}

	log.Infof("Queue size found: %d", queueSize)
	log.Infof("Nb running app found: %d", nbRunningApp)
	if queueSize > params.threshold && params.pendingJobs < params.maxPendingJob {
		log.Info("Hummmm... a Ti'Punch ^^")
		params.pendingJobs += 1
		err := connectors.Punch()
		if err != nil {
			log.Errorf("Unable to allocate resources %s", err)
		}
		params.pendingJobs -= 1
	} else if queueSize == 0 && nbRunningApp == 0 {
		connectors.QuitPunch()
	}
}

func RunForever() {
	for {
		go run()
		time.Sleep(1 * time.Second)
	}
}

func main() {
	log.Info("Starting Bebida optimizer service")
	params = Parameters{threshold: 1, pendingJobs: 0, maxPendingJob: 1}
	log.Infof("Parameters: %+v\n", params)
	RunForever()
}
