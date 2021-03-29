// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"database/sql"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
)

var accountSelect sq.SelectBuilder

func init() {
	accountSelect = sq.
		Select("Account.ID", "Provider", "Provisioner", "ProviderMetadataRaw", "AccountMetadataRaw",
			"State", "CreateAt", "DeleteAt",
			"APISecurityLock", "LockAcquiredBy", "LockAcquiredAt").
		From("Account")
}

// RawAccountMetadata is the raw byte metadata for a account.
type RawAccountMetadata struct {
	ProviderMetadataRaw []byte
	AccountMetadataRaw  []byte
}

type rawAccount struct {
	*model.Account
	*RawAccountMetadata
}

type rawAccounts []*rawAccount

func buildRawMetadata(account *model.Account) (*RawAccountMetadata, error) {
	providerMetadataJSON, err := json.Marshal(account.ProviderMetadataAWS)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal ProviderMetadataAWS")
	}

	accountMetadataJSON, err := json.Marshal(account.AccountMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal AccountMetadata")
	}

	return &RawAccountMetadata{
		ProviderMetadataRaw: providerMetadataJSON,
		AccountMetadataRaw:  accountMetadataJSON,
	}, nil
}

func (r *rawAccount) toAccount() (*model.Account, error) {
	var err error
	r.Account.ProviderMetadataAWS, err = model.NewAWSMetadata(r.ProviderMetadataRaw)
	if err != nil {
		return nil, err
	}

	r.Account.AccountMetadata, err = model.NewAccountMetadata(r.AccountMetadataRaw)
	if err != nil {
		return nil, err
	}

	return r.Account, nil
}

func (rc *rawAccounts) toAccounts() ([]*model.Account, error) {
	var accounts []*model.Account
	for _, rawAccount := range *rc {
		account, err := rawAccount.toAccount()
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetAccount fetches the given account by id.
func (sqlStore *SQLStore) GetAccount(id string) (*model.Account, error) {
	var rawAccount rawAccount
	err := sqlStore.getBuilder(sqlStore.db, &rawAccount, accountSelect.Where("ID = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get account by id")
	}

	return rawAccount.toAccount()
}

// GetAccounts fetches the given page of created accounts. The first page is 0.
func (sqlStore *SQLStore) GetAccounts(filter *model.AccountFilter) ([]*model.Account, error) {
	builder := accountSelect.
		OrderBy("CreateAt ASC")
	builder = sqlStore.applyAccountsFilter(builder, filter)

	var rawAccounts rawAccounts
	err := sqlStore.selectBuilder(sqlStore.db, &rawAccounts, builder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for accounts")
	}

	return rawAccounts.toAccounts()
}

func (sqlStore *SQLStore) applyAccountsFilter(builder sq.SelectBuilder, filter *model.AccountFilter) sq.SelectBuilder {
	if filter.PerPage != model.AllPerPage {
		builder = builder.
			Limit(uint64(filter.PerPage)).
			Offset(uint64(filter.Page * filter.PerPage))
	}

	if !filter.IncludeDeleted {
		builder = builder.Where("DeleteAt = 0")
	}

	return builder
}

// GetUnlockedAccountsPendingWork returns an unlocked account in a pending state.
func (sqlStore *SQLStore) GetUnlockedAccountsPendingWork() ([]*model.Account, error) {
	builder := accountSelect.
		Where(sq.Eq{
			"State": model.AllAccountStatesPendingWork,
		}).
		Where("LockAcquiredAt = 0").
		OrderBy("CreateAt ASC")

	var rawAccounts rawAccounts
	err := sqlStore.selectBuilder(sqlStore.db, &rawAccounts, builder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for accounts")
	}

	return rawAccounts.toAccounts()
}

// CreateAccount records the given account to the database, assigning it a unique ID.
func (sqlStore *SQLStore) CreateAccount(account *model.Account) error {
	tx, err := sqlStore.beginTransaction(sqlStore.db)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.RollbackUnlessCommitted()

	if err = sqlStore.createAccount(tx, account); err != nil {
		return errors.Wrap(err, "failed to create account")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit the transaction")
	}

	return nil
}

// createAccount records the given account to the database, assigning it a unique ID.
func (sqlStore *SQLStore) createAccount(execer execer, account *model.Account) error {
	account.ID = model.NewID()
	account.CreateAt = GetMillis()

	rawMetadata, err := buildRawMetadata(account)
	if err != nil {
		return errors.Wrap(err, "unable to build raw account metadata")
	}

	if _, err = sqlStore.execBuilder(execer, sq.
		Insert("Account").
		SetMap(map[string]interface{}{
			"ID":                  account.ID,
			"State":               account.State,
			"Provider":            account.Provider,
			"ProviderMetadataRaw": rawMetadata.ProviderMetadataRaw,
			"Provisioner":         account.Provisioner,
			"AccountMetadataRaw":  rawMetadata.AccountMetadataRaw,
			"CreateAt":            account.CreateAt,
			"DeleteAt":            account.DeleteAt,
			"APISecurityLock":     account.APISecurityLock,
			"LockAcquiredBy":      nil,
			"LockAcquiredAt":      0,
		}),
	); err != nil {
		return errors.Wrap(err, "failed to create account")
	}

	return nil
}

// UpdateAccount updates the given account in the database.
func (sqlStore *SQLStore) UpdateAccount(account *model.Account) error {
	rawMetadata, err := buildRawMetadata(account)
	if err != nil {
		return errors.Wrap(err, "unable to build raw account metadata")
	}

	if _, err = sqlStore.execBuilder(sqlStore.db, sq.
		Update("Account").
		SetMap(map[string]interface{}{
			"State":               account.State,
			"Provider":            account.Provider,
			"ProviderMetadataRaw": rawMetadata.ProviderMetadataRaw,
			"Provisioner":         account.Provisioner,
			"AccountMetadataRaw":  rawMetadata.AccountMetadataRaw,
		}).
		Where("ID = ?", account.ID),
	); err != nil {
		return errors.Wrap(err, "failed to update account")
	}

	return nil
}

// DeleteAccount marks the given account as deleted, but does not remove the record from the
// database.
func (sqlStore *SQLStore) DeleteAccount(id string) error {
	_, err := sqlStore.execBuilder(sqlStore.db, sq.
		Update("Account").
		Set("DeleteAt", GetMillis()).
		Where("ID = ?", id).
		Where("DeleteAt = 0"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to mark account as deleted")
	}

	return nil
}

// LockAccount marks the account as locked for exclusive use by the caller.
func (sqlStore *SQLStore) LockAccount(accountID, lockerID string) (bool, error) {
	return sqlStore.lockRows("Account", []string{accountID}, lockerID)
}

// UnlockAccount releases a lock previously acquired against a caller.
func (sqlStore *SQLStore) UnlockAccount(accountID, lockerID string, force bool) (bool, error) {
	return sqlStore.unlockRows("Account", []string{accountID}, lockerID, force)
}

// LockAccountAPI locks updates to the account from the API.
func (sqlStore *SQLStore) LockAccountAPI(accountID string) error {
	return sqlStore.setAccountAPILock(accountID, true)
}

// UnlockAccountAPI unlocks updates to the account from the API.
func (sqlStore *SQLStore) UnlockAccountAPI(accountID string) error {
	return sqlStore.setAccountAPILock(accountID, false)
}

func (sqlStore *SQLStore) setAccountAPILock(accountID string, lock bool) error {
	_, err := sqlStore.execBuilder(sqlStore.db, sq.
		Update("Account").
		Set("APISecurityLock", lock).
		Where("ID = ?", accountID),
	)
	if err != nil {
		return errors.Wrap(err, "failed to store account API lock")
	}

	return nil
}
