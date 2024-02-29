package connectors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPunch(t *testing.T) {

	ExecuteCommand = func(cmd string) (string, error) { return "\nOAR_JOB_ID=1234\nNot relevant", nil }
	jobId, _ := OAR{}.Punch(1, 10, time.Time{})

	assert.Equal(t, "1234", jobId, "Job id should be the same")
}
