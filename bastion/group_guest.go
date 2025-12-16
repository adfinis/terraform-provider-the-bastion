// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

import (
	"encoding/json"
	"fmt"
	"time"
)

type GroupGuestAccess ACL

// GroupListGuestAccesses lists all guest accesses from a group.
func (c *Client) GroupListGuestAccesses(group, account string) ([]*GroupGuestAccess, error) {
	response, err := c.executeCommand("groupListGuestAccesses", "--group", group, "--account", account)
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var accesses []*GroupGuestAccess
	if err := json.Unmarshal(valueBytes, &accesses); err != nil {
		return nil, err
	}

	return accesses, nil
}

// GroupAddGuestAccessOptions represents options for adding a guest access to a group.
type GroupAddGuestAccessOptions struct {
	TTL          string
	Comment      string
	Protocol     string
	ProxyOptions *ProxyOptions
	RemotePort   *int
}

func (g *GroupAddGuestAccessOptions) validate() error {
	if g.ProxyOptions != nil {
		if err := g.ProxyOptions.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GroupAddGuestAccessOptions) toArgs() []string {
	args := []string{}
	if g.TTL != "" {
		args = append(args, "--ttl", g.TTL)
	}
	if g.Comment != "" {
		args = append(args, "--comment", fmt.Sprintf("%q", g.Comment))
	}
	if g.Protocol != "" {
		args = append(args, "--protocol", g.Protocol)
	}
	if g.ProxyOptions != nil {
		args = append(args, g.ProxyOptions.toArgs()...)
	}
	if g.RemotePort != nil {
		args = append(args, "--remote-port", fmt.Sprintf("%d", *g.RemotePort))
	}
	return args
}

// GroupAddGuestAccess adds a guest access to a group.
func (c *Client) GroupAddGuestAccess(group, account, host, port, user string, options *GroupAddGuestAccessOptions) error {
	args := []string{"--group", group, "--account", account, "--host", host, "--port", fmt.Sprintf("%q", port)}
	if user != "" {
		args = append(args, "--user", fmt.Sprintf("%q", user))
	}
	if options != nil {
		if err := options.validate(); err != nil {
			return err
		}
		args = append(args, options.toArgs()...)
	}
	response, err := c.executeCommand("groupAddGuestAccess", args...)
	if err != nil {
		return err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return err
	}

	var access GroupGuestAccess
	if err := json.Unmarshal(valueBytes, &access); err != nil {
		return err
	}
	return nil
}

// GroupDelGuestAccess removes a guest access from a group.
func (c *Client) GroupDelGuestAccess(group, account, host, port, user, protocol string, proxyOpts *ProxyOptions, remotePort *int64) error {
	args := []string{"--group", group, "--account", account, "--host", host, "--port", fmt.Sprintf("%q", port)}
	if user != "" {
		args = append(args, "--user", fmt.Sprintf("%q", user))
	}
	if protocol != "" {
		args = append(args, "--protocol", protocol)
	}
	if proxyOpts != nil {
		if err := proxyOpts.validate(); err != nil {
			return err
		}
		args = append(args, proxyOpts.toArgs()...)
	}
	if remotePort != nil {
		args = append(args, "--remote-port", fmt.Sprintf("%d", *remotePort))
	}
	_, err := c.executeCommand("groupDelGuestAccess", args...)
	if err != nil {
		return err
	}

	// Ugly workaround because the bastion seems to fail if we call GroupDelGuestAccess too frequently
	time.Sleep(100 * time.Millisecond)

	return nil
}
