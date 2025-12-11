// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

import "encoding/json"

type AccountAccess struct {
	AccessType string  `json:"type"` // "personal", "group" or "group-guest"
	Group      *string `json:"group"`
	ACL        []ACL   `json:"acl"`
}

type ACL struct {
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
	RemotePort    *Port   `json:"remotePort"`
	LocalPort     *Port   `json:"localPort"`
}

// AccountListAccesses lists all accesses for an account.
func (c *Client) AccountListAccesses(account string) ([]*AccountAccess, error) {
	response, err := c.executeCommand("accountListAccesses", "--account", account)
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var accesses []*AccountAccess
	if err := json.Unmarshal(valueBytes, &accesses); err != nil {
		return nil, err
	}

	return accesses, nil
}
