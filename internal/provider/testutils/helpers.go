// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

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

func GetGroupKeyFingerprint(groupName string) (string, error) {
	group, err := TestBastionClient.GroupInfo(groupName)
	if err != nil {
		return "", err
	}

	// get the first available fingerprint
	for _, key := range group.Keys {
		if key.Fingerprint != "" {
			return key.Fingerprint, nil
		}
	}

	return "", nil
}

func GrantAccountCommand(account, command string) error {
	return TestBastionClient.AccountGrantCommand(account, command)
}

func RevokeAccountCommand(account, command string) error {
	return TestBastionClient.AccountRevokeCommand(account, command)
}

func CreateGroupServerAccess(group, ip, port, user string) error {
	_, err := TestBastionClient.GroupAddServer(group, ip, port, user, &bastion.GroupAddServerOptions{
		Force: true,
	})
	return err
}

func CreateGroupServerAccessWithProtocol(group, ip, port, protocol string) error {
	opts := &bastion.GroupAddServerOptions{
		Protocol: protocol,
		Force:    true,
	}
	_, err := TestBastionClient.GroupAddServer(group, ip, port, "", opts)
	return err
}

func DeleteGroupServerAccess(group, ip, port, user string) error {
	return TestBastionClient.GroupDelServer(group, ip, port, user, "", nil, nil)
}

func DeleteGroupServerAccessWithProtocol(group, ip, port, protocol string) error {
	return TestBastionClient.GroupDelServer(group, ip, port, "", protocol, nil, nil)
}
