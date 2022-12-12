package main

import (
	"testing"

	"github.com/RyaxTech/bebida-optimization-service/connectors"
)

func TestSSH(t *testing.T) {
	out, err := connectors.ExecuteCommand("echo toto")
	if err != nil {
		t.Error(err)
	}
	if out != "toto\n" {
		t.Errorf("Expected return was toto, got: %s.", out)
	}
}
