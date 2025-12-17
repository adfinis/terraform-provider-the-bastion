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

// WithSSHAgentAuth returns an SSH agent authentication method.
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

// WithPrivateKeyAuth returns a private key authentication method.
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

// WithPrivateKeyAuthWithPassphrase returns a private key authentication method with passphrase support.
func WithPrivateKeyAuthWithPassphrase(privateKey string, passphrase string) SSHAuthMethod {
	return func() (ssh.AuthMethod, error) {
		return getPrivateKeyAuthWithPassphrase(privateKey, passphrase)
	}
}

func getPrivateKeyAuthWithPassphrase(privateKey string, passphrase string) (ssh.AuthMethod, error) {
	var signer ssh.Signer
	var err error

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey([]byte(privateKey))
	}

	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

// WithPrivateKeyFileAuth returns a private key file authentication method.
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

// WithPrivateKeyFileAuthWithPassphrase returns a private key file authentication method with passphrase support.
func WithPrivateKeyFileAuthWithPassphrase(keyPath string, passphrase string) SSHAuthMethod {
	return func() (ssh.AuthMethod, error) {
		return getPrivateKeyFileAuthWithPassphrase(keyPath, passphrase)
	}
}

func getPrivateKeyFileAuthWithPassphrase(keyPath string, passphrase string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return getPrivateKeyAuthWithPassphrase(string(key), passphrase)
}
