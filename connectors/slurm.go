package connectors

import (
	"fmt"
	"strings"
	"time"

	"github.com/RyaxTech/bebida-shaker/connectors/exec"
	"github.com/RyaxTech/bebida-shaker/connectors/utils"
	"github.com/apex/log"
)

type SLURM struct{}

func (SLURM) Punch(nbCpuPerJob int, jobDuration time.Duration, deadline time.Time) (string, error) {
	randomSuffix := utils.RandomString(8)
	deadlineOption := ""
	if !deadline.IsZero() {
		deadlineOption = fmt.Sprintf("--begin=%s", deadline.Format(time.RFC3339))
	}
	cmd := fmt.Sprintf("sbatch --parsable --job-name BEBIDA_NOOP_%s -n %d --time %d %s sleep %d", randomSuffix, nbCpuPerJob, int(jobDuration.Minutes()), deadlineOption, int(jobDuration.Seconds()))
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
