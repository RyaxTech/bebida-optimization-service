package exec

import (
	"bytes"
	"encoding/base64"
	"os"

	"github.com/apex/log"
	"golang.org/x/crypto/ssh"
)

func ExecuteCommand(cmd string) (string, error) {

	// From file
	//key, err := os.ReadFile("pkey.tmp")
	//if err != nil {
	//	log.Fatalf("unable to read private key: %v", err)
	//}

	user := os.Getenv("BEBIDA_SSH_USER")
	hostname := os.Getenv("BEBIDA_SSH_HOSTNAME")
	port := os.Getenv("BEBIDA_SSH_PORT")
	connectionUrl := hostname + ":" + port
	// from base64 encoded env var
	keyBase64 := os.Getenv("BEBIDA_SSH_PKEY")
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		log.Fatalf("unable to decode private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
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
	var b bytes.Buffer
	session.Stdout = &b
	log.Infof("Running SSH on host %s with command: %s", hostname, cmd)
	if err := session.Run(cmd); err != nil {
		log.Error("Failed to run: " + err.Error())
		return "", err
	}
	log.Infof("Completed SSH on host %s with command: %s\nOUTPUTS: %s", hostname, cmd, b.String())
	return b.String(), nil
}
