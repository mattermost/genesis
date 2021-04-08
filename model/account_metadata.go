// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import "encoding/json"

// AccountMetadata is the provider metadata stored in a model.Account.
type AccountMetadata struct {
	Provision bool
	Subnet    string
}

// NewAccountMetadata creates an instance of AccountMetadata given the raw provider metadata.
func NewAccountMetadata(metadataBytes []byte) (*AccountMetadata, error) {
	if metadataBytes == nil || string(metadataBytes) == "null" {
		// TODO: remove "null" check after sqlite is gone.
		return nil, nil
	}

	var accountMetadata AccountMetadata
	err := json.Unmarshal(metadataBytes, &accountMetadata)
	if err != nil {
		return nil, err
	}

	return &accountMetadata, nil
}
