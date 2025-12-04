// Copyright (c) Adfinis
// SPDX-License-Identifier: GPL-3.0-or-later

package bastion

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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
