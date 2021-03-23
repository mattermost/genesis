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

// AddParentSubnetRequest specifies the parameters for a new parent subnet.
type AddParentSubnetRequest struct {
	CIDR       string `json:"cidr,omitempty"`
	SplitRange int    `json:"splitRange,omitempty"`
}

// SetDefaults sets the default values for a parent subnet create request.
func (request *AddParentSubnetRequest) SetDefaults() {
}

// Validate validates the values of a parent subnet create request.
func (request *AddParentSubnetRequest) Validate() error {
	if request.CIDR == "" {
		return errors.New("Parent CIDR cannot be empty")
	}

	return nil
}

// NewAddParentSubnetRequestFromReader will create a AddParentSubnetRequest from an
// io.Reader with JSON data.
func NewAddParentSubnetRequestFromReader(reader io.Reader) (*AddParentSubnetRequest, error) {
	var addParentSubnetRequest AddParentSubnetRequest
	err := json.NewDecoder(reader).Decode(&addParentSubnetRequest)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "failed to decode add parent subnet request")
	}

	addParentSubnetRequest.SetDefaults()
	err = addParentSubnetRequest.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "add parent subnet request failed validation")
	}

	return &addParentSubnetRequest, nil
}

// GetParentSubnetsRequest describes the parameters to request a list of parent subnets.
type GetParentSubnetsRequest struct {
	Page    int
	PerPage int
}

// ApplyToURL modifies the given url to include query string parameters for the request.
func (request *GetParentSubnetsRequest) ApplyToURL(u *url.URL) {
	q := u.Query()
	q.Add("page", strconv.Itoa(request.Page))
	q.Add("per_page", strconv.Itoa(request.PerPage))
	u.RawQuery = q.Encode()
}
