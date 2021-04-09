// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"encoding/json"
	"io"
)

// Account represents an AWS account.
type Account struct {
	ID                  string
	State               string
	Provider            string
	ProviderMetadataAWS *AWSMetadata
	AccountMetadata     *AccountMetadata
	Provisioner         string
	CreateAt            int64
	DeleteAt            int64
	APISecurityLock     bool
	LockAcquiredBy      *string
	LockAcquiredAt      int64
}

// AccountCreation stores information neeeded for account creation.
type AccountCreation struct {
	SSOUserEmail          string
	SSOFirstName          string
	SSOLastName           string
	ManagedOU             string
	ControlTowerRole      string
	ControlTowerAccountID string
}

// AccountProvision stores information neeeded for account provision.
type AccountProvision struct {
	StateBucket          string
	TransitGatewayID     string
	Environment          string
	TransitGatewayRoutes string
	TeleportCIDR         string
	CncCIDRs             string
	BindServerIPs        string
	ResourceShareID      string
	CoreAccountID        string
}

// Clone returns a deep copy the account.
func (a *Account) Clone() (*Account, error) {
	var clone Account
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}

	return &clone, nil
}

// AccountFromReader decodes a json-encoded account from the given io.Reader.
func AccountFromReader(reader io.Reader) (*Account, error) {
	account := Account{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&account)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &account, nil
}

// AccountsFromReader decodes a json-encoded list of accounts from the given io.Reader.
func AccountsFromReader(reader io.Reader) ([]*Account, error) {
	accounts := []*Account{}
	decoder := json.NewDecoder(reader)

	err := decoder.Decode(&accounts)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return accounts, nil
}

// AccountFilter describes the parameters used to constrain a set of accounts.
type AccountFilter struct {
	Page           int
	PerPage        int
	IncludeDeleted bool
}
