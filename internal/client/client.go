// Copyright (c) Christopher Barnes <christopher@barnes.biz>
// SPDX-License-Identifier: MPL-2.0

package roger

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
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

func NewClient(host, port string) (*Client, error) {
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 || p > 65535 {
		return nil, fmt.Errorf("invalid port: %q", port)
	}

	return &Client{Host: host, Port: p}, nil
}

func (c *Client) resolveFQDN() (string, error) {
	ips, err := net.LookupIP(c.Host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve IP for hostname %s: %w", c.Host, err)
	}

	for _, ip := range ips {
		if ip.To4() == nil {
			continue
		}
		ptrs, err := net.LookupAddr(ip.String())
		if err != nil {
			return "", fmt.Errorf("reverse lookup failed for IP %s: %w", ip, err)
		}
		if len(ptrs) > 0 {
			return strings.TrimSuffix(ptrs[0], "."), nil
		}
	}
	return "", fmt.Errorf("no valid IPv4 PTR record found for host %s", c.Host)
}

func (c *Client) CreateState(hostname, message, appstate string) (*State, error) {
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/", fqdn, c.Port)

	payload := fmt.Sprintf(`{"hostname": "%s", "message": "%s", "appstate": "%s"}`,
		hostname, message, appstate)

	cmd := exec.Command("curl", "-s",
		"--negotiate", "-u", ":",
		"-X", "POST",
		"-H", "Content-Type: application/json",
		"-H", "Accept: application/json",
		"-d", payload,
		"-w", "%{http_code}", // append http code to output
		url,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("curl POST failed: %v\nOutput: %s", err, out)
	}

	statusCodeStr := string(out[len(out)-3:])
	body := out[:len(out)-3]

	if statusCodeStr == "201" {
		return c.GetState(hostname)
	}

	var state State
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, fmt.Errorf("failed to parse POST response: %w\nRaw: %s", err, body)
	}

	return &state, nil
}

func (c *Client) GetState(hostname string) (*State, error) {
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s", fqdn, c.Port, hostname)
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
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s", fqdn, c.Port, hostname)
	cmd := exec.Command("curl", "-s", "--negotiate", "-u", ":", "-X", "DELETE",
		"-w", "%{http_code}", "-o", "/dev/stdout", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("curl failed: %w; output: %s", err, output)
	}
	return err
}

func (c *Client) UpdateState(hostname, message, appstate string) (*State, error) {
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s", fqdn, c.Port, hostname)

	payload := fmt.Sprintf(`{"hostname": "%s", "message": "%s", "appstate": "%s"}`,
		hostname, message, appstate)

	cmd := exec.Command("curl", "-s",
		"--negotiate", "-u", ":",
		"-X", "PUT",
		"-H", "Content-Type: application/json",
		"-H", "Accept: application/json",
		"-d", payload,
		url,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("curl PUT failed: %v\nOutput: %s", err, out)
	}

	var state State
	if err := json.Unmarshal(out, &state); err != nil {
		return nil, fmt.Errorf("failed to parse PUT response: %w\nRaw: %s", err, out)
	}

	return &state, nil
}