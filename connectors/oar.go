package connectors

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RyaxTech/bebida-shaker/connectors/exec"
	"github.com/RyaxTech/bebida-shaker/connectors/utils"
	"github.com/apex/log"
)

type OAR struct{}

var ExecuteCommand = exec.ExecuteCommand

func (OAR) Punch(nbCpuPerJob int, jobDuration time.Duration, deadline time.Time) (string, error) {
	// TODO put this in a config file (or env var)
	randomSuffix := utils.RandomString(8)
	oarsubOpts := fmt.Sprintf("--name BEBIDA_NOOP_%s -l nodes=%d,walltime=00:%d:00 \"sleep %d\"", randomSuffix, nbCpuPerJob, int(jobDuration.Minutes()), int(jobDuration.Seconds()))
	if !deadline.IsZero() {
		oarsubOpts = fmt.Sprintf("-r '%s' %s", deadline.Add(-jobDuration).Format("2006-01-02 15:04:05"), oarsubOpts)
	}

	// FIXME: user1 is hardcoded here, maybe we should use the right user for Bebida directly ass SSH level...
	cmd := fmt.Sprintf("su user1 --command 'oarsub %s'", oarsubOpts)
	out, err := ExecuteCommand(cmd)
	log.Infof("Punch command output: %s", string(out))

	// Find the job ID
	jobReg := regexp.MustCompile("OAR_JOB_ID=([0-9]+)")
	jobId := jobReg.FindStringSubmatch(out)[1]

	if err != nil {
		log.Errorf("Unable to submit job: %s", err)
		return "", err
	}

	return jobId, nil
}

func (OAR) QuitPunch(jobID string) error {
	cmd := fmt.Sprintf("oardel %s", jobID)
	out, err := ExecuteCommand(cmd)
	if err != nil {
		log.Errorf("Unable to delete job: %s", err)
		return err
	}

	log.Infof("Quit Punch command output: %s", string(out))
	return nil
}

func (oar OAR) QuitAllPunch() error {
	// get OAR job ID from the name
	cmd := string("oarstat --json | jq '.[] | select(.name | match(\"BEBIDA_NOOP\")) | .id' -r)")
	out, err := ExecuteCommand(cmd)
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

func (OAR) Refill(nbResources int) error {
	var quotaResource int
	if nbResources != -1 {
		// Apply quota on the server by changing the file content. It's reloaded for every scheduling round.
		cmd := string("oarnodes --json | jq '. | length'")
		out, err := ExecuteCommand(cmd)
		if err != nil {
			log.Errorf("Unable to list bebida jobs: %s", err)
			return err
		}
		totalResourceStr := strings.TrimSuffix(out, "\n")
		totalResources, err := strconv.Atoi(totalResourceStr)
		if err != nil {
			log.Errorf("Unable to parse number of resources: %s", err)
			return err
		}
		quotaResource = totalResources - nbResources
		if quotaResource < 0 {
			log.Errorf("Error while computing quota. Quota is negative: %d", quotaResource)
			return nil
		}
	} else {
		quotaResource = -1
	}
	// quotas format. Use * for all in, Use -1 in values for "no limit":
	// "<Queue>, <project>, <job_type>, <user>": [<Maximum used resources>, <Max running job>, <Max resource per hours>]
	quota := fmt.Sprintf("{\"quotas\": \"*,*,*,*\": [-1, %d, -1]}", quotaResource)
	cmd := fmt.Sprintf("echo '%s' > /etc/oar/quotas.json", quota)
	_, err := ExecuteCommand(cmd)
	if err != nil {
		log.Errorf("Unable to list bebida jobs: %s", err)
		return err
	}
	return nil
}
