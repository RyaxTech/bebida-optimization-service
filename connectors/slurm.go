package connectors

import (
	"fmt"
	"strings"

	"github.com/RyaxTech/bebida-shaker/connectors/exec"
	"github.com/RyaxTech/bebida-shaker/connectors/utils"
	"github.com/apex/log"
)

type SLURM struct{}

func (SLURM) Punch(nbCpuPerJob int, jobDurationInSeconds int) (string, error) {
	randomSuffix := utils.RandomString(8)
	cmd := fmt.Sprintf("sbatch --parsable --job-name BEBIDA_NOOP_%s -n %d sleep %d", randomSuffix, nbCpuPerJob, jobDurationInSeconds)
	out, err := exec.ExecuteCommand(cmd)
	if err != nil {
		log.Errorf("Unable to submit job: %s", err)
		return "", err
	}
	// Get job id
	jobID := strings.Split(out, ";")[0]

	log.Infof("Punch command output: %s", string(out))
	return jobID, nil
}

func (SLURM) QuitPunch(jobID string) error {
	cmd := fmt.Sprintf("scancel %s", jobID)
	out, err := exec.ExecuteCommand(cmd)
	if err != nil {
		log.Errorf("Unable to cancel job: %s", err)
		return err
	}

	log.Infof("Quit Punch command output: %s", string(out))
	return nil
}

func (SLURM) QuitAllPunch() error {
	log.Error("Not implemented!")
	return nil
}

func (SLURM) Refill(nbNodes int) error {
	log.Error("Not implemented!")
	return nil
}
