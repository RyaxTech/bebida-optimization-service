package connectors

import (
	"fmt"
	"strings"

	"github.com/RyaxTech/bebida-shaker/connectors/exec"
	"github.com/apex/log"
)

type OAR struct{}

func (OAR) Punch(nbCpuPerJob int, jobDurationInSeconds int) (string, error) {
	// TODO put this in a config file (or env var)
	cmd := fmt.Sprintf("oarsub --name BEBIDA_NOOP -l cores=%d sleep %d | grep OAR_JOB_ID | cut -d'=' -f2", nbCpuPerJob, jobDurationInSeconds)
	out, err := exec.ExecuteCommand(cmd)

	if err != nil {
		log.Errorf("Unable to submit job: %s", err)
		return "", err
	}

	log.Infof("Punch command output: %s", string(out))
	return out, nil
}

func (OAR) QuitPunch(jobID string) error {
	cmd := fmt.Sprintf("oardel %s", jobID)
	out, err := exec.ExecuteCommand(cmd)
	if err != nil {
		log.Errorf("Unable to delete job: %s", err)
		return err
	}

	log.Infof("Quit Punch command output: %s", string(out))
	return nil
}

func (oar OAR) QuitAllPunch(oarJobId string) error {
	// get OAR job ID from the name
	cmd := string("oarstat --json | jq '.[] | select(.name | match(\"BEBIDA_NOOP\")) | .id' -r)")
	out, err := exec.ExecuteCommand(cmd)
	if err != nil {
		log.Errorf("Unable to list bebida jobs: %s", err)
		return err
	}
	for _, oarJobID := range strings.Split(out, "\n") {
		if oarJobID != "" {
			err := oar.QuitPunch(oarJobID)
			if err != nil {
				log.Errorf("Unable to delete job: %s", err)
			}
		}
	}
	log.Infof("Quit Punch command output: %s", string(out))
	return nil
}

func (OAR) Refill(nbNodes int) error {
	log.Error("Not implemented!")
	return nil
}
