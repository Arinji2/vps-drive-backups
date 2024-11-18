package server

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

func SSHIntoServer() {
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
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	fmt.Println("Connected via SSH successfully")
	fmt.Println("Starting SSH Permissions Testing")

	err = session.Run("touch test.txt")
	if err == nil {
		fmt.Println("This user has access to create files. Please make the backups user readonly")

		err = session.Run("rm test.txt")
		if err == nil {
			log.Fatal("This user has access to delete files. Please make the backups user readonly")
		} else {
			log.Fatal("This file was not able to be deleted. Please delete the file 'test.txt' ")
		}

	}

	var stdout bytes.Buffer
	session.Stdout = &stdout
	session.Run("groups | grep -q sudo; echo $?")
	if (stdout.String()) == "0\n" {
		fmt.Println("This user has sudo access. Please make the backups user readonly")
	}

	fmt.Printf("\n\n\n\n\n PERMISSIONS TESTING COMPLETED \n\n\n\n\n")

}
