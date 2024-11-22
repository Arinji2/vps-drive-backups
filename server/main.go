package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func connectToServer() (*ssh.Session, *ssh.Client) {
	sshIP := os.Getenv("SSH_IP")
	sshUser := os.Getenv("SSH_USER")
	sshPwd := os.Getenv("SSH_PWD")

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(sshPwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", sshIP+":22", sshConfig)
	if err != nil {
		panic(err)
	}

	session, err := client.NewSession()
	if err != nil {
		panic(err)
	}

	return session, client
}

func SSHIntoServer() {
	session, client := connectToServer()
	defer client.Close()
	defer session.Close()

	fmt.Println("Connected via SSH successfully")

}

func GetBackup(vpsLocation string) io.Reader {
	session, client := connectToServer()
	defer session.Close()
	defer client.Close()

	// Run the remote commands to create and encode the tar.gz
	commands := []string{
		fmt.Sprintf("cd %s", vpsLocation),
		"tar -cf - --transform='s|^\\./||' . 2>/dev/null | gzip -9",
	}
	command := strings.Join(commands, " && ")

	// Capture the output of the command
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	// Start the command execution
	if err := session.Start(command); err != nil {
		log.Fatalf("Failed to run remote command: %v", err)
	}

	// Return the raw gzip data as an io.Reader
	return stdout
}
