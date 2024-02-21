package connectors

type HPCConnector interface {
	// Connector to HPC cluster applying the Shaker algorithm
	//
	// Punch a hole in the schedule and returns the JobID
	Punch(nbCpuPerJob int, jobDurationInSeconds int) (string, error)
	// Cancel the Punch job
	QuitPunch(jobID string) error
	// Cancel all Punch job
	QuitAllPunch() error
	// Increase or decrease the number of node reserved for the BDA workload
	Refill(nbNode int) error
}
