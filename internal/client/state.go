// Copyright (c) Christopher Barnes <christopher.barnes@cern.ch>
// SPDX-License-Identifier: GPL-3.0-or-later

package roger

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type State struct {
	AppAlarmed      bool   `json:"app_alarmed"`
	AppState        string `json:"appstate"`
	Expires         string `json:"expires"`
	ExpiresDT       string `json:"expires_dt"`
	Hostname        string `json:"hostname"`
	HWAlarmed       bool   `json:"hw_alarmed"`
	Message         string `json:"message"`
	NCAlarmed       bool   `json:"nc_alarmed"`
	OSAlarmed       bool   `json:"os_alarmed"`
	UpdatedTime     string `json:"update_time"`
	UpdatedTimeDT   string `json:"update_time_dt"`
	UpdatedBy       string `json:"updated_by"`
	UpdatedByPuppet bool   `json:"updated_by_puppet"`
}

func (c *Client) CreateState(hostname, message, appstate string) (*State, error) {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/", c.Host, c.Port)
	payload, _ := json.Marshal(map[string]string{
		"hostname": hostname,
		"message":  message,
		"appstate": appstate,
	})

	body, status, err := c.doRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}

	if status == http.StatusCreated {
		return c.GetState(hostname)
	}

	var state State
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &state, nil
}

func (c *Client) GetState(hostname string) (*State, error) {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s/", c.Host, c.Port, hostname)

	body, _, err := c.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &state, nil
}

func (c *Client) UpdateState(hostname, message, appstate string) (*State, error) {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s/", c.Host, c.Port, hostname)
	payload, _ := json.Marshal(map[string]string{
		"hostname": hostname,
		"message":  message,
		"appstate": appstate,
	})

	body, _, err := c.doRequest(http.MethodPut, url, payload)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &state, nil
}

func (c *Client) DeleteState(hostname string) error {
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s/", c.Host, c.Port, hostname)
	_, _, err := c.doRequest(http.MethodDelete, url, nil)
	return err
}
