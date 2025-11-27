// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package bastion

import (
	"encoding/json"
	"fmt"
	"time"
)

// Account represents a Bastion account.
type Account struct {
	// Account is the name of the account.
	Account string `json:"account"`
}

func (c *Client) AccountInfo(name string) (*Account, error) {
	response, err := c.executeCommand("accountInfo", "--account", name)
	if err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(response.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response value: %w", err)
	}

	var account Account
	if err := json.Unmarshal(valueBytes, &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account info: %w", err)
	}

	return &account, nil
}

// UIDOptions holds the uid options for creating an account.
type UIDOptions struct {
	auto bool
	uid  uint
}

// UIDOpt defines a function type for setting UID options.
type UIDOpt func(*UIDOptions)

// WithAutoUID sets the option to automatically assign a UID.
func WithAutoUID() UIDOpt {
	return func(o *UIDOptions) {
		o.auto = true
	}
}

// WithSpecificUID sets a specific UID for the account.
func WithSpecificUID(uid uint) UIDOpt {
	return func(o *UIDOptions) {
		o.uid = uid
	}
}

// CreateAccountOptions holds options for creating a Bastion account.
type CreateAccountOptions struct {
	AlwaysActive    bool
	OshOnly         bool
	MaxInactiveDays uint
	ImmutableKey    bool
	Comment         string
	PublicKey       string
	NoKey           bool
	TTL             time.Duration
}

func (c *CreateAccountOptions) validate() error {
	if c.NoKey && c.PublicKey != "" {
		return fmt.Errorf("cannot specify both NoKey and PublicKey")
	}
	return nil
}

func (c *CreateAccountOptions) toArgs() []string {
	args := []string{}

	if c.AlwaysActive {
		args = append(args, "--always-active")
	}
	if c.OshOnly {
		args = append(args, "--osh-only")
	}
	if c.MaxInactiveDays != 0 {
		args = append(args, "--max-inactive-days", fmt.Sprintf("%d", c.MaxInactiveDays))
	}
	if c.ImmutableKey {
		args = append(args, "--immutable-key")
	}
	if c.Comment != "" {
		args = append(args, "--comment", c.Comment)
	}
	if c.PublicKey != "" {
		args = append(args, "--public-key", c.PublicKey)
	}
	if c.NoKey {
		args = append(args, "--no-key")
	}
	if c.TTL != 0 {
		args = append(args, "--ttl", fmt.Sprintf("%ds", int(c.TTL.Seconds())))
	}

	return args
}

// CreateAccount creates a new Bastion account.
func (c *Client) CreateAccount(name string, uidOpt UIDOpt, createOpts *CreateAccountOptions) error {
	uidOption := &UIDOptions{}
	uidOpt(uidOption)

	if createOpts != nil {
		if err := createOpts.validate(); err != nil {
			return err
		}
	}

	args := []string{"--account", name}
	if uidOption.auto {
		args = append(args, "--uid-auto")
	}
	if uidOption.uid != 0 {
		args = append(args, "--uid", fmt.Sprintf("%d", uidOption.uid))
	}

	if createOpts != nil {
		args = append(args, createOpts.toArgs()...)
	}

	_, err := c.executeCommand("accountCreate", args...)
	if err != nil {
		return err
	}
	return nil
}

// YesNoBypass represents a three-state option: yes, no, or bypass.
type YesNoBypass string

const (
	YesNoBypassYes    YesNoBypass = "yes"
	YesNoBypassNo     YesNoBypass = "no"
	YesNoBypassBypass YesNoBypass = "bypass"
)

// YesNoDefault represents a three-state option: yes, no, or default.
type YesNoDefault string

const (
	YesNoDefaultYes     YesNoDefault = "yes"
	YesNoDefaultNo      YesNoDefault = "no"
	YesNoDefaultDefault YesNoDefault = "default"
)

// EgressStrictHostKeyCheckingPolicy represents the egress strict host key checking policies.
type EgressStrictHostKeyCheckingPolicy string

const (
	EgressStrictHostKeyCheckingYes      EgressStrictHostKeyCheckingPolicy = "yes"
	EgressStricHostKeyCheckingAcceptNew EgressStrictHostKeyCheckingPolicy = "accept-new"
	EgressStrictHostKeyCheckingNo       EgressStrictHostKeyCheckingPolicy = "no"
	EgressStrictHostKeyCheckingAsk      EgressStrictHostKeyCheckingPolicy = "ask"
	EgressStrictHostKeyCheckingDefault  EgressStrictHostKeyCheckingPolicy = "default"
	EgressStrictHostKeyCheckingBypass   EgressStrictHostKeyCheckingPolicy = "bypass"
)

// MFARequiredPolicy represents an MFA policies.
type MFARequiredPolicy string

const (
	MFARequiredPassword MFARequiredPolicy = "password"
	MFARequiredTOTP     MFARequiredPolicy = "totp"
	MFARequiredAny      MFARequiredPolicy = "any"
	MFARequiredNone     MFARequiredPolicy = "none"
)

// ModifyAccountOptions holds options for modifying a Bastion account.
type ModifyAccountOptions struct {
	PamAuthBypass               *bool
	MFAPasswordRequired         *YesNoBypass
	MFATOTPRequired             *YesNoBypass
	EgressStrictHostKeyChecking *EgressStrictHostKeyCheckingPolicy
	EgressSessionMultiplexing   *YesNoDefault
	PersonalEgressMFARequired   *MFARequiredPolicy
	AlwaysActive                *bool
	IdleIgnore                  *bool
	MaxInactiveDays             *int
	OshOnly                     *bool
	PubkeyAuthOptional          *bool
}

func (c *ModifyAccountOptions) toArgs() []string {
	args := []string{}

	if c.PamAuthBypass != nil {
		if *c.PamAuthBypass {
			args = append(args, "--pam-auth-bypass", "yes")
		} else {
			args = append(args, "--pam-auth-bypass", "no")
		}
	}
	if c.MFAPasswordRequired != nil {
		args = append(args, "--mfa-password-required", string(*c.MFAPasswordRequired))
	}
	if c.MFATOTPRequired != nil {
		args = append(args, "--mfa-totp-required", string(*c.MFATOTPRequired))
	}
	if c.EgressStrictHostKeyChecking != nil {
		args = append(args, "--egress-strict-host-key-checking", string(*c.EgressStrictHostKeyChecking))
	}
	if c.EgressSessionMultiplexing != nil {
		args = append(args, "--egress-session-multiplexing", string(*c.EgressSessionMultiplexing))
	}
	if c.PersonalEgressMFARequired != nil {
		args = append(args, "--personal-egress-mfa-required", string(*c.PersonalEgressMFARequired))
	}
	if c.AlwaysActive != nil {
		if *c.AlwaysActive {
			args = append(args, "--always-active", "yes")
		} else {
			args = append(args, "--always-active", "no")
		}
	}
	if c.IdleIgnore != nil {
		if *c.IdleIgnore {
			args = append(args, "--idle-ignore", "yes")
		} else {
			args = append(args, "--idle-ignore", "no")
		}
	}
	if c.MaxInactiveDays != nil {
		args = append(args, "--max-inactive-days", fmt.Sprintf("%d", *c.MaxInactiveDays))
	}
	if c.OshOnly != nil {
		if *c.OshOnly {
			args = append(args, "--osh-only", "yes")
		} else {
			args = append(args, "--osh-only", "no")
		}
	}
	if c.PubkeyAuthOptional != nil {
		if *c.PubkeyAuthOptional {
			args = append(args, "--pubkey-auth-optional", "yes")
		} else {
			args = append(args, "--pubkey-auth-optional", "no")
		}
	}

	return args
}

// ModifyAccount modifies an existing Bastion account.
func (c *Client) ModifyAccount(name string, modifyOpts *ModifyAccountOptions) error {
	if modifyOpts == nil {
		return fmt.Errorf("modify options cannot be nil")
	}

	args := []string{"--account", name}
	args = append(args, modifyOpts.toArgs()...)

	_, err := c.executeCommand("accountModify", args...)
	if err != nil {
		return err
	}
	return nil
}

// DeleteAccount deletes a Bastion account.
func (c *Client) DeleteAccount(name string) error {
	_, err := c.executeCommand("accountDelete", "--account", name, "--no-confirm")
	if err != nil {
		return err
	}
	return nil
}
