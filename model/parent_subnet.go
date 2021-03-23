// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"encoding/json"
	"io"
)

// ParentSubnet represents a parent subnet range.
type ParentSubnet struct {
	ID             string
	CIDR           string
	State          string
	SplitRange     int
	CreateAt       int64
	LockAcquiredBy *string
	LockAcquiredAt int64
}

// Clone returns a deep copy of the parent subnet.
func (c *ParentSubnet) Clone() *ParentSubnet {
	var clone ParentSubnet
	data, _ := json.Marshal(c)
	json.Unmarshal(data, &clone)

	return &clone
}

// ParentSubnetFromReader decodes a json-encoded parent subnet from the given io.Reader.
func ParentSubnetFromReader(reader io.Reader) (*ParentSubnet, error) {
	account := ParentSubnet{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&account)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &account, nil
}

// ParentSubnetsFromReader decodes a json-encoded list of parent subnets from the given io.Reader.
func ParentSubnetsFromReader(reader io.Reader) ([]*ParentSubnet, error) {
	parentSubnets := []*ParentSubnet{}
	decoder := json.NewDecoder(reader)

	err := decoder.Decode(&parentSubnets)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return parentSubnets, nil
}

// ParentSubnetFilter describes the parameters used to constrain a set of parent subnets.
type ParentSubnetFilter struct {
	Page    int
	PerPage int
}
