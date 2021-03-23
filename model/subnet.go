// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"encoding/json"
	"io"
)

// Subnet represents a parent subnet range.
type Subnet struct {
	ID             string
	CIDR           string
	Used           bool
	ParentSubnet   string
	SubnetMetadata *SubnetMetadata
	CreateAt       int64
	LockAcquiredBy *string
	LockAcquiredAt int64
}

// Clone returns a deep copy of the subnet.
func (c *Subnet) Clone() *Subnet {
	var clone Subnet
	data, _ := json.Marshal(c)
	json.Unmarshal(data, &clone)

	return &clone
}

// SubnetFromReader decodes a json-encoded subnet from the given io.Reader.
func SubnetFromReader(reader io.Reader) (*Subnet, error) {
	account := Subnet{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&account)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &account, nil
}

// SubnetsFromReader decodes a json-encoded list of subnets from the given io.Reader.
func SubnetsFromReader(reader io.Reader) ([]*Subnet, error) {
	subnets := []*Subnet{}
	decoder := json.NewDecoder(reader)

	err := decoder.Decode(&subnets)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return subnets, nil
}

// SubnetFilter describes the parameters used to constrain a set of subnets.
type SubnetFilter struct {
	Page    int
	PerPage int
	Free    bool
}
