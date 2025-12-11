// Copyright (c) Adfinis
// SPDX-License-Identifier: MPL-2.0

package bastion

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIResponse represents the standard API response from The Bastion.
type APIResponse struct {
	Command      string `json:"command"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Value        any    `json:"value"`
}

// Error implements the error interface for APIResponse.
func (e *APIResponse) Error() string {
	return fmt.Sprintf("Bastion API error [%s]: %s (command: %s)", e.ErrorCode, e.ErrorMessage, e.Command)
}

// executeCommand executes a command on The Bastion and returns the JSON response.
func (c *Client) executeCommand(command string, args ...string) (*APIResponse, error) {
	sshClient, err := c.sshClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}
	defer sshClient.Close() //nolint:errcheck

	// Build the command for The Bastion: --osh <command> <args> --json-greppable --quiet
	fullCommand := fmt.Sprintf("--osh %s %s --json-greppable --quiet", command, strings.Join(args, " "))
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close() //nolint:errcheck

	output, err := session.CombinedOutput(fullCommand)
	response, parseErr := parseJSONGreppableOutput(string(output))
	if parseErr != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to execute command: %w, output: %s", err, string(output))
		}
		return nil, fmt.Errorf("failed to parse response: %w, output: %s", parseErr, string(output))
	}

	if !response.isSuccess() {
		return nil, response
	}

	return response, nil
}

func (r *APIResponse) isSuccess() bool {
	return strings.HasPrefix(r.ErrorCode, "OK")
}

// parseJSONGreppableOutput parses the JSON output from --json-greppable format.
func parseJSONGreppableOutput(output string) (*APIResponse, error) {
	for line := range strings.SplitSeq(output, "\n") {
		if jsonData, ok := strings.CutPrefix(line, "JSON_OUTPUT="); ok {
			var response APIResponse
			if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
			}

			return &response, nil
		}
	}

	return nil, fmt.Errorf("no JSON output found in response")
}
