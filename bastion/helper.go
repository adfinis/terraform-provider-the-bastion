// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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

// PIVPolicy represents the PIV policy for account ingress keys.
type PIVPolicy string

const (
	PIVPolicyDefault PIVPolicy = "default"
	PIVPolicyEnforce PIVPolicy = "enforce"
	PIVPolicyNever   PIVPolicy = "never"
	PIVPolicyGrace   PIVPolicy = "grace"
)

// BoolFromInt is simple and works like this 1 => true, 0 => false.
type BoolFromInt bool

func (b *BoolFromInt) UnmarshalJSON(data []byte) error {
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*b = BoolFromInt(intVal != 0)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into BoolFromInt", string(data))
}

func (b BoolFromInt) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

func (b BoolFromInt) Bool() bool {
	return bool(b)
}

// Port is a helper to represent port which can be a string or int.
type Port struct {
	Number int
	raw    any
}

func NewPort(port string) *Port {
	return &Port{
		raw: port,
	}
}

func (p *Port) ValueString() string {
	switch v := p.raw.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return ""
	}
}

func (p *Port) ValueInt() int {
	switch v := p.raw.(type) {
	case int:
		return v
	case string:
		sv, _ := strconv.Atoi(v)
		return sv
	default:
		return 0
	}
}

func (p *Port) UnmarshalJSON(data []byte) error {
	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		p.Number = asInt
		p.raw = asInt
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		p.raw = asString
		return nil
	}

	return nil
}

func (p *Port) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.raw)
}
