// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
)

// GetAccountDTO fetches the given account by id with data from connected tables.
func (sqlStore *SQLStore) GetAccountDTO(id string) (*model.AccountDTO, error) {
	account, err := sqlStore.GetAccount(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account")
	}
	if account == nil {
		return nil, nil
	}

	return &model.AccountDTO{
		Account: account,
	}, nil
}

// GetAccountDTOs fetches the given page of accounts with data from connected tables. The first page is 0.
func (sqlStore *SQLStore) GetAccountDTOs(filter *model.AccountFilter) ([]*model.AccountDTO, error) {
	accounts, err := sqlStore.GetAccounts(filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	dtos := make([]*model.AccountDTO, 0, len(accounts))
	for _, c := range accounts {
		dtos = append(dtos, &model.AccountDTO{Account: c})
	}

	return dtos, nil
}
