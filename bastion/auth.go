// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHAuthMethod defines a function that returns an SSH authentication method.
type SSHAuthMethod func() (ssh.AuthMethod, error)

// WithSSHAgentAuth returns SSH agent authentication method.
func WithSSHAgentAuth() SSHAuthMethod {
	return func() (ssh.AuthMethod, error) {
		return getSSHAgentAuth()
	}
}

func getSSHAgentAuth() (ssh.AuthMethod, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers), nil
}

// WithPrivateKeyAuth returns private key authentication method.
func WithPrivateKeyAuth(privateKey string) SSHAuthMethod {
	return func() (ssh.AuthMethod, error) {
		return getPrivateKeyAuth(privateKey)
	}
}

func getPrivateKeyAuth(privateKey string) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

// WithPrivateKeyFileAuth returns private key file authentication method.
func WithPrivateKeyFileAuth(keyPath string) SSHAuthMethod {
	return func() (ssh.AuthMethod, error) {
		return getPrivateKeyFileAuth(keyPath)
	}
}

func getPrivateKeyFileAuth(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return getPrivateKeyAuth(string(key))
}
