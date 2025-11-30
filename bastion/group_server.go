// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package bastion

import (
	"encoding/json"
	"fmt"
)

// GroupServer represents a Bastion group server access.
type GroupServer struct {
	IP            string  `json:"ip"`
	Port          *Port   `json:"port"`
	User          *string `json:"user"`
	ProxyIP       *string `json:"proxyIp"`
	ProxyPort     *Port   `json:"proxyPort"`
	ProxyUser     *string `json:"proxyUser"`
	Comment       *string `json:"comment"`
	UserComment   *string `json:"userComment"`
	ForcePassword *string `json:"forcePassword"`
	ForceKey      *string `json:"forceKey"`
	Protocol      *string `json:"protocol"`
	ReverseDNS    *string `json:"reverseDns"`
	AddedBy       string  `json:"addedBy"`
	AddedDate     string  `json:"addedDate"`
	Expiry        *int    `json:"expiry"`
}

// GroupListServers lists all accesses from a group.
func (c *Client) GroupListServers(name string) ([]*GroupServer, error) {
	response, err := c.executeCommand("groupListServers", "--group", name)
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var servers []*GroupServer
	if err := json.Unmarshal(valueBytes, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}

// GroupAddServerOptions represents options for adding a server access to a group.
type GroupAddServerOptions struct {
	Force         bool
	ForceKey      string
	ForcePassword string
	TTL           string
	Comment       string
	Protocol      string
	ProxyOptions  *ProxyOptions
}

// ProxyOptions respresents proxy options for adding an access.
type ProxyOptions struct {
	ProxyHost string
	ProxyPort string
	ProxyUser string
}

func (p *ProxyOptions) validate() error {
	if p.ProxyHost == "" {
		return ErrProxyMissingHost
	}
	if p.ProxyPort == "" {
		return ErrProxyMissingPort
	}
	if p.ProxyUser == "" {
		return ErrProxyMissingUser
	}
	return nil
}

func (p *ProxyOptions) toArgs() []string {
	var args []string
	args = append(args, "--proxy-host", p.ProxyHost)
	args = append(args, "--proxy-port", p.ProxyPort)
	args = append(args, "--proxy-user", fmt.Sprintf("%q", p.ProxyUser))
	return args
}

func (g *GroupAddServerOptions) validate() error {
	if g.ProxyOptions != nil {
		if err := g.ProxyOptions.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (g *GroupAddServerOptions) toArgs() []string {
	var args []string
	if g.Force {
		args = append(args, "--force")
	}
	if g.ForceKey != "" {
		args = append(args, "--force-key", g.ForceKey)
	}
	if g.ForcePassword != "" {
		args = append(args, "--force-password", g.ForcePassword)
	}
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
	return args
}

// GroupAddServer adds a server access to a group.
func (c *Client) GroupAddServer(group, host, port, user string, options *GroupAddServerOptions) (*GroupServer, error) {
	args := []string{"--group", group, "--host", host, "--port", fmt.Sprintf("%q", port)}
	if user != "" {
		args = append(args, "--user", fmt.Sprintf("%q", user))
	}
	if options != nil {
		if err := options.validate(); err != nil {
			return nil, err
		}
		args = append(args, options.toArgs()...)
	}
	response, err := c.executeCommand("groupAddServer", args...)
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var server GroupServer
	if err := json.Unmarshal(valueBytes, &server); err != nil {
		return nil, err
	}
	return &server, nil
}

// GroupDelServer removes a server access from a group.
func (c *Client) GroupDelServer(group, host, port, user, protocol string, proxyOpts *ProxyOptions) error {
	args := []string{"--group", group, "--host", host, "--port", fmt.Sprintf("%q", port)}
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
	_, err := c.executeCommand("groupDelServer", args...)
	return err
}
