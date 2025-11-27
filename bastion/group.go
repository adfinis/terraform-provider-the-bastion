// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package bastion

import (
	"encoding/json"
	"fmt"
)

// Group represents a Bastion group.
type Group struct {
	Group           string             `json:"group"`
	Inactive        []string           `json:"inactive"`
	Guests          []string           `json:"guests"`
	Owners          []string           `json:"owners"`
	Members         []string           `json:"members"`
	Gatekeepers     []string           `json:"gatekeepers"`
	ACLKeepers      []string           `json:"aclkeepers"`
	GuestAccesses   []string           `json:"guest_accesses"`
	Keys            map[string]Key     `json:"keys"`
	MFARequired     *MFARequiredPolicy `json:"mfa_required"`
	IdleLockTimeout *string            `json:"idle_lock_timeout"`
	IdleKillTimeout *string            `json:"idle_kill_timeout"`
	GuestTtlLimit   *string            `json:"guest_ttl_limit"`
}

// GroupInfo returns information about a Bastion group.
func (c *Client) GroupInfo(name string) (*Group, error) {
	response, err := c.executeCommand("groupInfo", "--group", name)
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(valueBytes, &group); err != nil {
		return nil, err
	}

	return &group, nil
}

// CreateGroup creates a new Bastion group.
func (c *Client) CreateGroup(name, owner string, keyAlgo KeyAlgo) (*Group, error) {
	algo, size := keyAlgo.AlgoAndSize()
	response, err := c.executeCommand("groupCreate", "--group", name, "--owner", owner, "--algo", algo, "--size", fmt.Sprintf("%d", size))
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(valueBytes, &group); err != nil {
		return nil, err
	}

	return &group, nil
}

// GroupModifyOptions holds options for modifying a Bastion group.
type GroupModifyOptions struct {
	MFARequired     *MFARequiredPolicy
	IdleLockTimeout *string
	IdleKillTimeout *string
	GuestTtlLimit   *string
}

func (g *GroupModifyOptions) toArgs() []string {
	var args []string
	if g.MFARequired != nil {
		args = append(args, "--mfa-required", string(*g.MFARequired))
	}
	if g.IdleLockTimeout != nil {
		args = append(args, "--idle-lock-timeout", *g.IdleLockTimeout)
	}
	if g.IdleKillTimeout != nil {
		args = append(args, "--idle-kill-timeout", *g.IdleKillTimeout)
	}
	if g.GuestTtlLimit != nil {
		args = append(args, "--guest-ttl-limit", *g.GuestTtlLimit)
	}
	return args
}

// ModifyGroup modifies a Bastion group.
func (c *Client) ModifyGroup(name string, modifyOpts *GroupModifyOptions) error {
	args := []string{"--group", name}
	if modifyOpts != nil {
		args = append(args, modifyOpts.toArgs()...)
	}
	_, err := c.executeCommand("groupModify", args...)
	return err
}

// DeleteGroup deletes a Bastion group.
// This is a restricted command that allows deletion of any group.
func (c *Client) DeleteGroup(name string) error {
	_, err := c.executeCommand("groupDelete", "--group", name, "--no-confirm")
	return err
}

// DestroyGroup deletes a Bastion group.
// This command can be used by group owners to delete their own groups.
func (c *Client) DestroyGroup(name string) error {
	_, err := c.executeCommand("groupDestroy", "--group", name, "--no-confirm")
	return err
}

// GroupAddOwner adds an owner to a Bastion group.
func (c *Client) GroupAddOwner(group, account string) error {
	_, err := c.executeCommand("groupAddOwner", "--group", group, "--account", account)
	return err
}

// GroupAddGatekeeper adds a gatekeeper to a Bastion group.
func (c *Client) GroupAddGatekeeper(group, account string) error {
	_, err := c.executeCommand("groupAddGatekeeper", "--group", group, "--account", account)
	return err
}

// GroupAddACLKeeper adds an ACL keeper to a Bastion group.
func (c *Client) GroupAddACLKeeper(group, account string) error {
	_, err := c.executeCommand("groupAddAclkeeper", "--group", group, "--account", account)
	return err
}

// GroupAddMember adds a member to a Bastion group.
func (c *Client) GroupAddMember(group, account string) error {
	_, err := c.executeCommand("groupAddMember", "--group", group, "--account", account)
	return err
}

// GroupRemoveOwner removes an owner from a Bastion group.
func (c *Client) GroupRemoveOwner(group, account string) error {
	_, err := c.executeCommand("groupDelOwner", "--group", group, "--account", account)
	return err
}

// GroupRemoveGatekeeper removes a gatekeeper from a Bastion group.
func (c *Client) GroupRemoveGatekeeper(group, account string) error {
	_, err := c.executeCommand("groupDelGatekeeper", "--group", group, "--account", account)
	return err
}

// GroupRemoveACLKeeper removes an ACL keeper from a Bastion group.
func (c *Client) GroupRemoveACLKeeper(group, account string) error {
	_, err := c.executeCommand("groupDelAclkeeper", "--group", group, "--account", account)
	return err
}

// GroupRemoveMember removes a member from a Bastion group.
func (c *Client) GroupRemoveMember(group, account string) error {
	_, err := c.executeCommand("groupDelMember", "--group", group, "--account", account)
	return err
}
