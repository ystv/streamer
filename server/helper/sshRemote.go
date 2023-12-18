package helper

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// ConnectToHostPassword is a general function to ssh to a remote server, any code execution is handled outside this function
func ConnectToHostPassword(host, username, password string, verbose bool) (*ssh.Client, *ssh.Session, error) {
	if verbose {
		fmt.Println("Connect To Host Password called")
	}
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		err = client.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	return client, session, nil
}

func RunCommandOnHost(host, username, password, command string, verbose bool) (string, error) {
	var client *ssh.Client
	var session *ssh.Session
	var err error
	//if forwarderAuth == "PEM" {
	//	client, session, err = connectToHostPEM(forwarder, forwarderUsername, forwarderPrivateKey, forwarderPassphrase)
	//} else if forwarderAuth == "PASS" {
	client, session, err = ConnectToHostPassword(host, username, password, verbose)
	//}
	if err != nil {
		return "", fmt.Errorf("error connecting to host: %w", err)
	}
	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("error running command on remote host: %w", err)
	}
	err = client.Close()
	if err != nil {
		return "", fmt.Errorf("error closing connection to remote host: %w", err)
	}
	return string(output), nil
}
