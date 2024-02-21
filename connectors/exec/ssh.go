package exec

import (
	"encoding/base64"
	"os"

	"github.com/apex/log"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	user      string
	hostname  string
	port      string
	keyBase64 string
}

var setup = setupFromEnv

func setupFromEnv() SSHConfig {
	user := os.Getenv("BEBIDA_SSH_USER")
	hostname := os.Getenv("BEBIDA_SSH_HOSTNAME")
	port := os.Getenv("BEBIDA_SSH_PORT")
	// from base64 encoded env var
	keyBase64 := os.Getenv("BEBIDA_SSH_PKEY")
	return SSHConfig{user: user, hostname: hostname, port: port, keyBase64: keyBase64}
}

func ExecuteCommand(cmd string) (string, error) {
	sshConfig := setup()
	connectionUrl := sshConfig.hostname + ":" + sshConfig.port

	key, err := base64.StdEncoding.DecodeString(sshConfig.keyBase64)
	if err != nil {
		log.Fatalf("unable to decode private key: %v", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: sshConfig.user,
		Auth: []ssh.AuthMethod{
			// Use the PublicKeys method for remote authentication.
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", connectionUrl, config)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}
	defer client.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	log.Infof("Running SSH on host '%s' with command: %s", sshConfig.hostname, cmd)
	if out, err := session.CombinedOutput(cmd); err != nil {
		log.Error("Failed to run: " + err.Error())
		return string(out), err
	} else {
		log.Infof("Completed SSH on host %s with command: %s\nOUTPUTS: %s", sshConfig.hostname, cmd, out)
		return string(out), nil
	}
}
