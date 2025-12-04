// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package testutils

import (
	"os"

	"github.com/adfinis/terraform-provider-bastion/bastion"
)

var (
	TestBastionClient *bastion.Client
	SSHPrivateKey     []byte
	SSHPublicKey      []byte
)

func init() {
	var err error
	SSHPrivateKey, err = os.ReadFile("../../ssh-keys/id_ed25519")
	if err != nil {
		panic("failed to read SSH private key: " + err.Error())
	}
	SSHPublicKey, err = os.ReadFile("../../ssh-keys/id_ed25519.pub")
	if err != nil {
		panic("failed to read SSH public key: " + err.Error())
	}
	client, err := bastion.New(&bastion.Config{
		Host:                  "localhost",
		Port:                  2222,
		Username:              "bastionadmin",
		StrictHostKeyChecking: false,
	}, bastion.WithPrivateKeyAuth(string(SSHPrivateKey)))
	if err != nil {
		panic("failed to create test bastion client: " + err.Error())
	}
	TestBastionClient = client
}

func CreateAccounts(names ...string) (err error) {
	for _, name := range names {
		if err = CreateAccount(name); err != nil {
			return
		}
	}
	return
}

func CreateAccount(name string) error {
	createOpts := &bastion.CreateAccountOptions{
		PublicKey: string(SSHPublicKey),
	}
	return TestBastionClient.CreateAccount(name, bastion.WithAutoUID(), createOpts)
}

func DeleteAccounts(names ...string) (err error) {
	for _, name := range names {
		if err = DeleteAccount(name); err != nil {
			return
		}
	}
	return
}

func DeleteAccount(name string) error {
	return TestBastionClient.DeleteAccount(name)
}

func CreateGroups(owner string, keyAlgo bastion.KeyAlgo, names ...string) (err error) {
	for _, name := range names {
		if err = CreateGroup(name, owner, keyAlgo); err != nil {
			return
		}
	}
	return
}

func CreateGroup(name, owner string, keyAlgo bastion.KeyAlgo) error {
	_, err := TestBastionClient.CreateGroup(name, owner, keyAlgo)
	return err
}

func DeleteGroups(names ...string) (err error) {
	for _, name := range names {
		if err = DeleteGroup(name); err != nil {
			return
		}
	}
	return
}

func DeleteGroup(name string) error {
	if err := TestBastionClient.DestroyGroup(name); err == nil {
		return nil
	}
	return TestBastionClient.DeleteGroup(name)
}
