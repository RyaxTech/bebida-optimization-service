package connectors

import (
	"fmt"

	"github.com/apex/log"
)

type SlurmConfig struct {
	nbCpuPerJob          int
	jobDurationInSeconds int
}

func Punch() error {
	// TODO put this in a config file (or env var)
	config := SlurmConfig{nbCpuPerJob: 1, jobDurationInSeconds: 900}

	cmd := fmt.Sprintf("srun --job-name BEBIDA_NOOP -n %d sleep %d", config.nbCpuPerJob, config.jobDurationInSeconds)
	out, err := ExecuteCommand(cmd)

	if err != nil {
		log.Errorf("Unable to submit job: %s", err)
		return err
	}

	log.Infof("Punch command output: %s", string(out))
	return nil
}
