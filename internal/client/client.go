// Copyright (c) Christopher Barnes <christopher@barnes.biz>
// SPDX-License-Identifier: MPL-2.0

package roger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	Host       string
	Port       int
}

type State struct {
	AppAlarmed      string `json:"app_alarmed"`
	AppState        string `json:"appstate"`
	Expires         string `json:"expires"`
	ExpiresDT       string `json:"expires_dt"`
	Hostname        string `json:"hostname"`
	HWAlarmed       string `json:"hw_alarmed"`
	Message         string `json:"message"`
	NCAlarmed       string `json:"nc_alarmed"`
	OSAlarmed       string `json:"os_alarmed"`
	UpdatedTime     string `json:"update_time"`
	UpdatedTimeDT   string `json:"update_time_dt"`
	UpdatedBy       string `json:"updated_by"`
	UpdatedByPuppet string `json:"updated_by_puppet"`
}

func NewClient(host, port *string) (*Client, error) {
	c := &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Host:       "",
		Port:       8080,
	}

	if host != nil && *host != "" {
		c.Host = *host
	}

	if port != nil && *port != "" {
		p, err := strconv.Atoi(*port)
		if err != nil || p <= 0 || p > 65535 {
			return nil, fmt.Errorf("invalid port: %q", *port)
		}
		c.Port = p
	}

	return c, nil
}

func (c *Client) CreateState(hostname, message, appstate string) (*State, error) {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s", c.Host, c.Port, hostname)
	cmd := exec.Command("curl", "-s", "--negotiate", "-u", ":", "-X", "POST",
		"-H", "Content-Type: application/json", url,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("curl failed: %v\nOutput: %s", err, out)
	}

	var state State
	if err := json.Unmarshal(out, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state response: %w\nRaw: %s", err, out)
	}

	return &state, nil
}

func (c *Client) GetState(hostname string) (*State, error) {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s", c.Host, c.Port, hostname)
	cmd := exec.Command("curl", "-s", "--negotiate", "-u", ":", "-X", "GET", url)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("curl failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute curl: %w", err)
	}

	var state State
	if err := json.Unmarshal(output, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &state, nil
}

func (c *Client) DeleteState(hostname string) error {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s", c.Host, c.Port, hostname)
	cmd := exec.Command("curl", "-s", "--negotiate", "-u", ":", "-X", "DELETE",
		"-w", "%{http_code}", "-o", "/dev/stdout", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("curl failed: %w; output: %s", err, output)
	}
	return err
}

func (c *Client) UpdateState(hostname, message, appstate string) (*State, error) {
	statePtr, err := c.GetState(hostname)
	if err != nil {
		return nil, err
	}
	return statePtr, nil
}
