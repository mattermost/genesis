// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"testing"
	"time"

	"github.com/mattermost/genesis/internal/testlib"
	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/require"
)

func TestAccounts(t *testing.T) {
	t.Run("get unknown account", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		account, err := sqlStore.GetAccount("unknown")
		require.NoError(t, err)
		require.Nil(t, account)
	})

	t.Run("get accounts", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		account1 := &model.Account{
			Provider:            "aws",
			Provisioner:         "genesis",
			ProviderMetadataAWS: &model.AWSMetadata{ServiceCatalogProductID: "prod-12345"},
			AccountMetadata:     &model.AccountMetadata{Provision: true},
			State:               model.AccountStateCreationRequested,
		}

		err := sqlStore.CreateAccount(account1)
		require.NoError(t, err)

		actualAccount1, err := sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, account1, actualAccount1)

		actualAccounts, err := sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 0, IncludeDeleted: false})
		require.NoError(t, err)
		require.Empty(t, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 1, IncludeDeleted: false})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 10, IncludeDeleted: false})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 1, IncludeDeleted: true})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 10, IncludeDeleted: true})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{PerPage: model.AllPerPage, IncludeDeleted: true})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)
	})

	t.Run("update accounts", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		account1 := &model.Account{
			Provider:    "aws",
			Provisioner: "kops",
			ProviderMetadataAWS: &model.AWSMetadata{
				ServiceCatalogProductID: "prod-12345",
				AWSAccountID:            "12345678",
			},
			AccountMetadata: &model.AccountMetadata{Provision: true},
			State:           model.AccountStateCreationRequested,
		}

		err := sqlStore.CreateAccount(account1)
		require.NoError(t, err)

		account1.Provider = "azure"
		account1.ProviderMetadataAWS = &model.AWSMetadata{
			ServiceCatalogProductID: "prod-12345",
			AWSAccountID:            "12345678",
		}
		account1.State = model.AccountStateDeletionRequested

		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		actualAccount1, err := sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, account1, actualAccount1)
	})

	t.Run("delete account", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		account1 := &model.Account{
			Provider:            "aws",
			Provisioner:         "genesis",
			ProviderMetadataAWS: &model.AWSMetadata{ServiceCatalogProductID: "prod-12345"},
			AccountMetadata:     &model.AccountMetadata{Provision: true},
			State:               model.AccountStateCreationRequested,
		}

		err := sqlStore.CreateAccount(account1)
		require.NoError(t, err)

		err = sqlStore.DeleteAccount(account1.ID)
		require.NoError(t, err)

		actualAccount1, err := sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.NotEqual(t, 0, actualAccount1.DeleteAt)
		account1.DeleteAt = actualAccount1.DeleteAt
		require.Equal(t, account1, actualAccount1)

		actualAccounts, err := sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 0, IncludeDeleted: false})
		require.NoError(t, err)
		require.Empty(t, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 1, IncludeDeleted: true})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)

		actualAccounts, err = sqlStore.GetAccounts(&model.AccountFilter{Page: 0, PerPage: 10, IncludeDeleted: true})
		require.NoError(t, err)
		require.Equal(t, []*model.Account{account1}, actualAccounts)

		time.Sleep(1 * time.Millisecond)

		// Deleting again shouldn't change timestamp
		err = sqlStore.DeleteAccount(account1.ID)
		require.NoError(t, err)

		actualAccount1, err = sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, account1, actualAccount1)

	})
}

func TestGetUnlockedAccountsPendingWork(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := MakeTestSQLStore(t, logger)

	creationRequestedAccount := &model.Account{
		State: model.AccountStateCreationRequested,
	}
	err := sqlStore.CreateAccount(creationRequestedAccount)
	require.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	provisioningRequestedAccount := &model.Account{
		State: model.AccountStateProvisioningRequested,
	}
	err = sqlStore.CreateAccount(provisioningRequestedAccount)
	require.NoError(t, err)

	time.Sleep(1 * time.Millisecond)

	deletionRequestedAccount := &model.Account{
		State: model.AccountStateDeletionRequested,
	}
	err = sqlStore.CreateAccount(deletionRequestedAccount)
	require.NoError(t, err)

	// Store accounts with states that should be ignored by GetUnlockedAccountsPendingWork()
	otherStates := []string{
		model.AccountStateCreationFailed,
		model.AccountStateProvisioningFailed,
		model.AccountStateDeletionFailed,
		model.AccountStateDeleted,
		model.AccountStateStable,
	}
	for _, otherState := range otherStates {
		err = sqlStore.CreateAccount(&model.Account{State: otherState})
		require.NoError(t, err)
	}

	accounts, err := sqlStore.GetUnlockedAccountsPendingWork()
	require.NoError(t, err)
	require.Equal(t, []*model.Account{creationRequestedAccount, provisioningRequestedAccount, deletionRequestedAccount}, accounts)

	lockerID := model.NewID()

	locked, err := sqlStore.LockAccount(creationRequestedAccount.ID, lockerID)
	require.NoError(t, err)
	require.True(t, locked)

	accounts, err = sqlStore.GetUnlockedAccountsPendingWork()
	require.NoError(t, err)
	require.Equal(t, []*model.Account{provisioningRequestedAccount, deletionRequestedAccount}, accounts)

	locked, err = sqlStore.LockAccount(provisioningRequestedAccount.ID, lockerID)
	require.NoError(t, err)
	require.True(t, locked)

	accounts, err = sqlStore.GetUnlockedAccountsPendingWork()
	require.NoError(t, err)
	require.Equal(t, []*model.Account{deletionRequestedAccount}, accounts)

	locked, err = sqlStore.LockAccount(deletionRequestedAccount.ID, lockerID)
	require.NoError(t, err)
	require.True(t, locked)

	accounts, err = sqlStore.GetUnlockedAccountsPendingWork()
	require.NoError(t, err)
	require.Empty(t, accounts)
}

func TestLockAccount(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := MakeTestSQLStore(t, logger)

	lockerID1 := model.NewID()
	lockerID2 := model.NewID()

	account1 := &model.Account{}
	err := sqlStore.CreateAccount(account1)
	require.NoError(t, err)

	account2 := &model.Account{}
	err = sqlStore.CreateAccount(account2)
	require.NoError(t, err)

	t.Run("accounts should start unlocked", func(t *testing.T) {
		account1, err = sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), account1.LockAcquiredAt)
		require.Nil(t, account1.LockAcquiredBy)

		account2, err = sqlStore.GetAccount(account2.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), account2.LockAcquiredAt)
		require.Nil(t, account2.LockAcquiredBy)
	})

	t.Run("lock an unlocked account", func(t *testing.T) {
		locked, err := sqlStore.LockAccount(account1.ID, lockerID1)
		require.NoError(t, err)
		require.True(t, locked)

		account1, err = sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.NotEqual(t, int64(0), account1.LockAcquiredAt)
		require.Equal(t, lockerID1, *account1.LockAcquiredBy)
	})

	t.Run("lock a previously locked account", func(t *testing.T) {
		t.Run("by the same locker", func(t *testing.T) {
			locked, err := sqlStore.LockAccount(account1.ID, lockerID1)
			require.NoError(t, err)
			require.False(t, locked)
		})

		t.Run("by a different locker", func(t *testing.T) {
			locked, err := sqlStore.LockAccount(account1.ID, lockerID2)
			require.NoError(t, err)
			require.False(t, locked)
		})
	})

	t.Run("lock a second account from a different locker", func(t *testing.T) {
		locked, err := sqlStore.LockAccount(account2.ID, lockerID2)
		require.NoError(t, err)
		require.True(t, locked)

		account2, err = sqlStore.GetAccount(account2.ID)
		require.NoError(t, err)
		require.NotEqual(t, int64(0), account2.LockAcquiredAt)
		require.Equal(t, lockerID2, *account2.LockAcquiredBy)
	})

	t.Run("unlock the first account", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockAccount(account1.ID, lockerID1, false)
		require.NoError(t, err)
		require.True(t, unlocked)

		account1, err = sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), account1.LockAcquiredAt)
		require.Nil(t, account1.LockAcquiredBy)
	})

	t.Run("unlock the first account again", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockAccount(account1.ID, lockerID1, false)
		require.NoError(t, err)
		require.False(t, unlocked)

		account1, err = sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), account1.LockAcquiredAt)
		require.Nil(t, account1.LockAcquiredBy)
	})

	t.Run("force unlock the first account again", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockAccount(account1.ID, lockerID1, true)
		require.NoError(t, err)
		require.False(t, unlocked)

		account1, err = sqlStore.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), account1.LockAcquiredAt)
		require.Nil(t, account1.LockAcquiredBy)
	})

	t.Run("unlock the second account from the wrong locker", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockAccount(account2.ID, lockerID1, false)
		require.NoError(t, err)
		require.False(t, unlocked)

		account2, err = sqlStore.GetAccount(account2.ID)
		require.NoError(t, err)
		require.NotEqual(t, int64(0), account2.LockAcquiredAt)
		require.Equal(t, lockerID2, *account2.LockAcquiredBy)
	})

	t.Run("force unlock the second account from the wrong locker", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockAccount(account2.ID, lockerID1, true)
		require.NoError(t, err)
		require.True(t, unlocked)

		account2, err = sqlStore.GetAccount(account2.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), account2.LockAcquiredAt)
		require.Nil(t, account2.LockAcquiredBy)
	})
}
