// Copyright (c) Christopher Barnes <christopher@barnes.biz>
// SPDX-License-Identifier: MPL-2.0

package roger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/credentials"
	"github.com/jcmturner/gokrb5/v8/spnego"
)

type Client struct {
	HTTPClient *spnego.Client
	Host       string
	Port       int
}

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


func loadKrb5Config() (*config.Config, error) {
	path := os.Getenv("KRB5_CONFIG")
	if path == "" {
		path = "/etc/krb5.conf" // fallback
	}
	return config.Load(path)
}

func loadCCache() (*credentials.CCache, error) {
	ccachePath := os.Getenv("KRB5CCNAME")
	if ccachePath == "" {
		return nil, fmt.Errorf("KRB5CCNAME environment variable not set")
	}
	
	ccachePath = strings.TrimPrefix(ccachePath, "FILE:")
	return credentials.LoadCCache(ccachePath)
}

func NewClient(host, port string) (*Client, error) {
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 || p > 65535 {
		return nil, fmt.Errorf("invalid port: %q", port)
	}

	krbConf, err := loadKrb5Config()
	if err != nil {
		return nil, fmt.Errorf("failed to load krb5.conf: %w", err)
	}
	
	ccache, err := loadCCache()
	if err != nil {
		return nil, fmt.Errorf("failed to load credential cache: %w", err)
	}
	
	krbClient, err := client.NewFromCCache(ccache, krbConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create kerberos client: %w", err)
	}

	httpClient := spnego.NewClient(krbClient, nil, "")

	return &Client{Host: host, Port: p, HTTPClient: httpClient}, nil
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

func (c *Client) doRequest(method, url string, payload []byte) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp.StatusCode, nil
}

func (c *Client) CreateState(hostname, message, appstate string) (*State, error) {
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/", fqdn, c.Port)
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
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s/", fqdn, c.Port, hostname)

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
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s/", fqdn, c.Port, hostname)
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
	fqdn, err := c.resolveFQDN()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s:%d/roger/v1/state/%s/", fqdn, c.Port, hostname)
	_, _, err = c.doRequest(http.MethodDelete, url, nil)
	return err
}
