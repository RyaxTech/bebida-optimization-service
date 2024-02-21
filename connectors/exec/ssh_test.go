package exec

import (
	"os"
	"testing"
)

func TestSSH(t *testing.T) {
	// Assuming that you have ssh on localhost on port 22 working
	setup = func() SSHConfig {return SSHConfig{user: os.Getenv("USER"), hostname: "localhost", port: "22", keyBase64: os.Getenv("BEBIDA_SSH_PKEY")}}
	out, err := ExecuteCommand("echo toto")
	if err != nil {
		t.Error(err)
	}
	if out != "toto\n" {
		t.Errorf("Expected return was toto, got: %s.", out)
	}
}
