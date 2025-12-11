// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

import (
	"encoding/json"
	"fmt"
)

type CreationInformation struct {
	Timestamp      int    `json:"timestamp"`
	Comment        string `json:"comment"`
	By             string `json:"by"`
	BastionVersion string `json:"bastion_version"`
}

type LastActivity struct {
	Timestamp int    `json:"timestamp"`
	Ago       string `json:"ago"`
}

type IngressPIVGrace struct {
	Enabled             BoolFromInt `json:"enabled"`
	ExpirationTimestamp int         `json:"expiration_timestamp"`
	SecondsRemaining    int         `json:"seconds_remaining"`
}

// Account represents a Bastion account.
type Account struct {
	Account                   string              `json:"account"`
	MFATOTPBypass             BoolFromInt         `json:"mfa_totp_bypass"`
	MFATOTPRequired           BoolFromInt         `json:"mfa_totp_required"`
	MFATOTPConfigured         BoolFromInt         `json:"mfa_totp_configured"`
	MFAPasswordBypass         BoolFromInt         `json:"mfa_password_bypass"`
	MFAPasswordRequired       BoolFromInt         `json:"mfa_password_required"`
	MFAPasswordConfigured     BoolFromInt         `json:"mfa_password_configured"`
	GlobalIngressPolicy       BoolFromInt         `json:"global_ingress_policy"`
	IsExpired                 BoolFromInt         `json:"is_expired"`
	PersonalEgressMFARequired MFARequiredPolicy   `json:"personal_egress_mfa_required"`
	CreationInformation       CreationInformation `json:"creation_information"`
	AllowedCommands           []string            `json:"allowed_commands"`
	IngressPIVPolicy          PIVPolicy           `json:"ingress_piv_policy"`
	IngressPIVEnforced        BoolFromInt         `json:"ingress_piv_enforced"`
	IngressPIVGrace           IngressPIVGrace     `json:"ingress_piv_grace"`
	CanConnect                BoolFromInt         `json:"can_connect"`
	AlreadySeenBefore         BoolFromInt         `json:"already_seen_before"`
	IsActive                  BoolFromInt         `json:"is_active"`
	AlwaysActive              BoolFromInt         `json:"always_active"`
	MaxInactiveDays           string              `json:"max_inactive_days"`
	IsFrozen                  BoolFromInt         `json:"is_frozen"`
	OshOnly                   BoolFromInt         `json:"osh_only"`
	IsAdmin                   BoolFromInt         `json:"is_admin"`
	IsSuperOwner              BoolFromInt         `json:"is_super_owner"`
	IsAuditor                 BoolFromInt         `json:"is_auditor"`
	IsTTLSet                  BoolFromInt         `json:"is_ttl_set"`
	IsTTLExpired              BoolFromInt         `json:"is_ttl_expired"`
	TTTLTimestamp             int                 `json:"ttl_timestamp"`
	IdleIgnore                BoolFromInt         `json:"idle_ignore"`
	PamAuthBypass             BoolFromInt         `json:"pam_auth_bypass"`
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
	TTL             int
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
		args = append(args, "--ttl", fmt.Sprintf("%ds", c.TTL))
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
	if c.MFAPasswordRequired != nil && *c.MFAPasswordRequired != "" {
		args = append(args, "--mfa-password-required", string(*c.MFAPasswordRequired))
	}
	if c.MFATOTPRequired != nil && *c.MFATOTPRequired != "" {
		args = append(args, "--mfa-totp-required", string(*c.MFATOTPRequired))
	}
	if c.EgressStrictHostKeyChecking != nil && *c.EgressStrictHostKeyChecking != "" {
		args = append(args, "--egress-strict-host-key-checking", string(*c.EgressStrictHostKeyChecking))
	}
	if c.EgressSessionMultiplexing != nil && *c.EgressSessionMultiplexing != "" {
		args = append(args, "--egress-session-multiplexing", string(*c.EgressSessionMultiplexing))
	}
	if c.PersonalEgressMFARequired != nil && *c.PersonalEgressMFARequired != "" {
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

// AccuntGrantCommand grants a command to a Bastion account.
func (c *Client) AccountGrantCommand(account, command string) error {
	_, err := c.executeCommand("accountGrantCommand", "--account", account, "--command", command)
	if err != nil {
		return err
	}
	return nil
}

// AccountRevokeCommand revokes a command from a Bastion account.
func (c *Client) AccountRevokeCommand(account, command string) error {
	_, err := c.executeCommand("accountRevokeCommand", "--account", account, "--command", command)
	if err != nil {
		return err
	}
	return nil
}

// AccountSetPIVPolicy sets the PIV policy for an account.
func (c *Client) AccountSetPIVPolicy(account string, policy PIVPolicy) error {
	if policy == PIVPolicyGrace {
		return fmt.Errorf("use AccountSetPIVGrace for grace policy")
	}

	_, err := c.executeCommand("accountPIV", "--account", account, "--policy", string(policy))
	if err != nil {
		return err
	}
	return nil
}

// AccountSetPIVGrace sets the PIV grace policy for an account with a TTL.
// The ttl parameter is in seconds.
func (c *Client) AccountSetPIVGrace(account string, ttl int) error {
	_, err := c.executeCommand("accountPIV", "--account", account, "--policy", "grace", "--ttl", fmt.Sprintf("%d", ttl))
	if err != nil {
		return err
	}
	return nil
}
