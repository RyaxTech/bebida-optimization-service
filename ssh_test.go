package main

import (
	"testing"

	"github.com/RyaxTech/bebida-shaker/connectors/exec"
)

func TestSSH(t *testing.T) {
	out, err := exec.ExecuteCommand("echo toto")
	if err != nil {
		t.Error(err)
	}
	if out != "toto\n" {
		t.Errorf("Expected return was toto, got: %s.", out)
	}
}
