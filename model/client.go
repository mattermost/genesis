// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client is the programmatic interface to the genesis server API.
type Client struct {
	address    string
	headers    map[string]string
	httpClient *http.Client
}

// NewClient creates a client to the genesis server at the given address.
func NewClient(address string) *Client {
	return &Client{
		address:    address,
		headers:    make(map[string]string),
		httpClient: &http.Client{},
	}
}

// NewClientWithHeaders creates a client to the genesis server at the given
// address and uses the provided headers.
func NewClientWithHeaders(address string, headers map[string]string) *Client {
	return &Client{
		address:    address,
		headers:    headers,
		httpClient: &http.Client{},
	}
}

// closeBody ensures the Body of an http.Response is properly closed.
func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
	}
}

func (c *Client) buildURL(urlPath string, args ...interface{}) string {
	return fmt.Sprintf("%s%s", c.address, fmt.Sprintf(urlPath, args...))
}

func (c *Client) doGet(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *Client) doPost(u string, request interface{}) (*http.Response, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) doPut(u string, request interface{}) (*http.Response, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}

	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *Client) doDelete(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	return c.httpClient.Do(req)
}

// CreateAccount requests the creation of an account from the configured genesis server.
func (c *Client) CreateAccount(request *CreateAccountRequest) (*Account, error) {
	resp, err := c.doPost(c.buildURL("/api/accounts"), request)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusAccepted:
		return AccountFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// RetryCreateAccount retries the creation of an account from the configured genesis server.
func (c *Client) RetryCreateAccount(accountID string) error {
	resp, err := c.doPost(c.buildURL("/api/account/%s", accountID), nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil

	default:
		return errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// ProvisionAccount provisions k8s operators and Helm charts on a
// account from the configured genesis server.
func (c *Client) ProvisionAccount(accountID string, request *ProvisionAccountRequest) (*Account, error) {
	resp, err := c.doPost(c.buildURL("/api/account/%s/provision", accountID), request)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusAccepted:
		return AccountFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetAccount fetches the specified account from the configured genesis server.
func (c *Client) GetAccount(accountID string) (*Account, error) {
	resp, err := c.doGet(c.buildURL("/api/account/%s", accountID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return AccountFromReader(resp.Body)

	case http.StatusNotFound:
		return nil, nil

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetAccounts fetches the list of accounts from the configured genesis server.
func (c *Client) GetAccounts(request *GetAccountsRequest) ([]*Account, error) {
	u, err := url.Parse(c.buildURL("/api/accounts"))
	if err != nil {
		return nil, err
	}

	request.ApplyToURL(u)

	resp, err := c.doGet(u.String())
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return AccountsFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// DeleteAccount deletes the given account and all resources contained therein.
func (c *Client) DeleteAccount(accountID string) error {
	resp, err := c.doDelete(c.buildURL("/api/account/%s", accountID))
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusAccepted:
		return nil

	default:
		return errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// CreateWebhook requests the creation of a webhook from the configured genesis server.
func (c *Client) CreateWebhook(request *CreateWebhookRequest) (*Webhook, error) {
	resp, err := c.doPost(c.buildURL("/api/webhooks"), request)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusAccepted:
		return WebhookFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetWebhook fetches the webhook from the configured genesis server.
func (c *Client) GetWebhook(webhookID string) (*Webhook, error) {
	resp, err := c.doGet(c.buildURL("/api/webhook/%s", webhookID))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return WebhookFromReader(resp.Body)

	case http.StatusNotFound:
		return nil, nil

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetWebhooks fetches the list of webhooks from the configured genesis server.
func (c *Client) GetWebhooks(request *GetWebhooksRequest) ([]*Webhook, error) {
	u, err := url.Parse(c.buildURL("/api/webhooks"))
	if err != nil {
		return nil, err
	}

	request.ApplyToURL(u)

	resp, err := c.doGet(u.String())
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return WebhooksFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// DeleteWebhook deletes the given webhook.
func (c *Client) DeleteWebhook(webhookID string) error {
	resp, err := c.doDelete(c.buildURL("/api/webhook/%s", webhookID))
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	default:
		return errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// LockAPIForAccount locks API changes for a given account.
func (c *Client) LockAPIForAccount(accountID string) error {
	return c.makeSecurityCall("account", accountID, "api", "lock")
}

// UnlockAPIForAccount unlocks API changes for a given account.
func (c *Client) UnlockAPIForAccount(accountID string) error {
	return c.makeSecurityCall("account", accountID, "api", "unlock")
}

func (c *Client) makeSecurityCall(resourceType, id, securityType, action string) error {
	resp, err := c.doPost(c.buildURL("/api/security/%s/%s/%s/%s", resourceType, id, securityType, action), nil)
	if err != nil {
		return err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	default:
		return errors.Errorf("failed with status code %d", resp.StatusCode)
	}

}

// AddParentSubnet requests the addition of a parent subnet from the configured genesis server.
func (c *Client) AddParentSubnet(request *AddParentSubnetRequest) (*ParentSubnet, error) {
	resp, err := c.doPost(c.buildURL("/api/subnets/parent"), request)
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusAccepted, http.StatusCreated:
		return ParentSubnetFromReader(resp.Body)
	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetParentSubnets fetches the list of parent subnets from the configured genesis server.
func (c *Client) GetParentSubnets(request *GetParentSubnetsRequest) ([]*ParentSubnet, error) {
	u, err := url.Parse(c.buildURL("/api/subnets/parent"))
	if err != nil {
		return nil, err
	}

	request.ApplyToURL(u)

	resp, err := c.doGet(u.String())
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return ParentSubnetsFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetParentSubnet fetches the specified parent subnet from the configured genesis server.
func (c *Client) GetParentSubnet(subnet string) (*ParentSubnet, error) {
	resp, err := c.doGet(c.buildURL("/api/subnet/parent/%s", subnet))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return ParentSubnetFromReader(resp.Body)

	case http.StatusNotFound:
		return nil, nil

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetSubnets fetches the list of subnets from the configured genesis server.
func (c *Client) GetSubnets(request *GetSubnetsRequest) ([]*Subnet, error) {
	u, err := url.Parse(c.buildURL("/api/subnets"))
	if err != nil {
		return nil, err
	}

	request.ApplyToURL(u)

	resp, err := c.doGet(u.String())
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return SubnetsFromReader(resp.Body)

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}

// GetSubnet fetches the specified subnet from the configured genesis server.
func (c *Client) GetSubnet(subnet string) (*Subnet, error) {
	resp, err := c.doGet(c.buildURL("/api/subnet/%s", subnet))
	if err != nil {
		return nil, err
	}
	defer closeBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
		return SubnetFromReader(resp.Body)

	case http.StatusNotFound:
		return nil, nil

	default:
		return nil, errors.Errorf("failed with status code %d", resp.StatusCode)
	}
}
