package connectors

type HPConnector interface {
	// Connector to HPC cluster applying the Shaker algorithm
	//
	// Punch a hole in the schedule and returns the JobID
	Punch(nbCpuPerJob int, jobDurationInSeconds int) (string, error)
	// Cancel the Punch job
	QuitPunch(jobID string) error
	// Increase or decrease the number of node reserved for the BDA workload
	Refill(nbNode int) error
}
