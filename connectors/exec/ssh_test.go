package exec

import (
	"testing"
)

func TestSSH(t *testing.T) {
	out, err := ExecuteCommand("echo toto")
	if err != nil {
		t.Error(err)
	}
	if out != "toto\n" {
		t.Errorf("Expected return was toto, got: %s.", out)
	}
}
