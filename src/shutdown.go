package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

// ShutdownPC connects via SSH and issues the shutdown command
func ShutdownPC(targetIP, user, keyPath string) error {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", targetIP, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Execute Windows shutdown command with 3 seconds delay to avoid immediate disconnect errors
	err = session.Run("shutdown /s /t 3")
	if err != nil {
		return fmt.Errorf("failed to run shutdown command: %v", err)
	}

	return nil
}
