// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"encoding/json"
	"io"
)

// AccountDTO represents account entity with connected data. DTO stands for Data Transfer Object.
type AccountDTO struct {
	*Account
}

// AccountDTOFromReader decodes a json-encoded account DTO from the given io.Reader.
func AccountDTOFromReader(reader io.Reader) (*AccountDTO, error) {
	accountDTO := AccountDTO{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&accountDTO)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &accountDTO, nil
}

// AccountDTOsFromReader decodes a json-encoded list of cluster DTOs from the given io.Reader.
func AccountDTOsFromReader(reader io.Reader) ([]*AccountDTO, error) {
	accountDTOs := []*AccountDTO{}
	decoder := json.NewDecoder(reader)

	err := decoder.Decode(&accountDTOs)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return accountDTOs, nil
}
