// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// CreateAccountRequest specifies the parameters for a new account.
type CreateAccountRequest struct {
	Provider                string `json:"provider,omitempty"`
	ServiceCatalogProductID string `json:"serviceCatalogProductID,omitempty"`
	Provision               bool   `json:"provision,omitempty"`
	APISecurityLock         bool   `json:"api-security-lock,omitempty"`
}

// SetDefaults sets the default values for an account create request.
func (request *CreateAccountRequest) SetDefaults() {
	if len(request.Provider) == 0 {
		request.Provider = ProviderAWS
	}
}

// Validate validates the values of an account create request.
func (request *CreateAccountRequest) Validate() error {
	if request.Provider != ProviderAWS {
		return errors.Errorf("unsupported provider %s", request.Provider)
	}

	if request.ServiceCatalogProductID == "" {
		return errors.New("Service Catalog Product ID cannot be empty")
	}

	return nil
}

// NewCreateAccountRequestFromReader will create a CreateAccountRequest from an
// io.Reader with JSON data.
func NewCreateAccountRequestFromReader(reader io.Reader) (*CreateAccountRequest, error) {
	var createAccountRequest CreateAccountRequest
	err := json.NewDecoder(reader).Decode(&createAccountRequest)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "failed to decode create account request")
	}

	createAccountRequest.SetDefaults()
	err = createAccountRequest.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "create account request failed validation")
	}

	return &createAccountRequest, nil
}

// GetAccountsRequest describes the parameters to request a list of accounts.
type GetAccountsRequest struct {
	Page           int
	PerPage        int
	IncludeDeleted bool
}

// ApplyToURL modifies the given url to include query string parameters for the request.
func (request *GetAccountsRequest) ApplyToURL(u *url.URL) {
	q := u.Query()
	q.Add("page", strconv.Itoa(request.Page))
	q.Add("per_page", strconv.Itoa(request.PerPage))
	if request.IncludeDeleted {
		q.Add("include_deleted", "true")
	}
	u.RawQuery = q.Encode()
}

// ProvisionAccountRequest contains metadata related to changing the installed account state.
type ProvisionAccountRequest struct {
}

// NewProvisionAccountRequestFromReader will create an UpdateAccountRequest from an io.Reader with JSON data.
func NewProvisionAccountRequestFromReader(reader io.Reader) (*ProvisionAccountRequest, error) {
	var provisionAccountRequest ProvisionAccountRequest
	err := json.NewDecoder(reader).Decode(&provisionAccountRequest)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "failed to decode provision account request")
	}
	return &provisionAccountRequest, nil
}
