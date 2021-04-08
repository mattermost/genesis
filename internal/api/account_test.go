// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/genesis/internal/api"
	"github.com/mattermost/genesis/internal/store"
	"github.com/mattermost/genesis/internal/testlib"
	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccounts(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := store.MakeTestSQLStore(t, logger)
	defer store.CloseConnection(t, sqlStore)

	router := mux.NewRouter()
	api.Register(router, &api.Context{
		Store:      sqlStore,
		Supervisor: &mockSupervisor{},
		Logger:     logger,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := model.NewClient(ts.URL)

	t.Run("unknown account", func(t *testing.T) {
		account, err := client.GetAccount(model.NewID())
		require.NoError(t, err)
		require.Nil(t, account)
	})

	t.Run("no accounts", func(t *testing.T) {
		accounts, err := client.GetAccounts(&model.GetAccountsRequest{
			Page:           0,
			PerPage:        10,
			IncludeDeleted: true,
		})
		require.NoError(t, err)
		require.Empty(t, accounts)
	})

	t.Run("get accounts", func(t *testing.T) {
		t.Run("invalid page", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/accounts?page=invalid&per_page=100", ts.URL))
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("invalid perPage", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/accounts?page=0&per_page=invalid", ts.URL))
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})

		t.Run("no paging parameters", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/accounts", ts.URL))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})

		t.Run("missing page", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/accounts?per_page=100", ts.URL))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})

		t.Run("missing perPage", func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/api/accounts?page=1", ts.URL))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})
	t.Run("accounts", func(t *testing.T) {
		account1, err := client.CreateAccount(&model.CreateAccountRequest{
			Provider:                model.ProviderAWS,
			ServiceCatalogProductID: "service-catalog-id",
			Provision:               true,
		})
		require.NoError(t, err)
		require.NotNil(t, account1)
		require.Equal(t, model.ProviderAWS, account1.Provider)

		actualAccount1, err := client.GetAccount(account1.ID)
		logger.Info(actualAccount1.ProviderMetadataAWS)
		require.NoError(t, err)
		require.Equal(t, account1.ID, actualAccount1.ID)
		require.Equal(t, model.ProviderAWS, actualAccount1.Provider)
		require.Equal(t, model.AccountStateCreationRequested, actualAccount1.State)
		require.Equal(t, "service-catalog-id", actualAccount1.ProviderMetadataAWS.ServiceCatalogProductID)
		require.Equal(t, true, actualAccount1.AccountMetadata.Provision)

		time.Sleep(1 * time.Millisecond)

		account2, err := client.CreateAccount(&model.CreateAccountRequest{
			Provider:                model.ProviderAWS,
			ServiceCatalogProductID: "service-catalog-id",
			Provision:               false,
		})
		require.NoError(t, err)
		require.NotNil(t, account2)
		require.Equal(t, model.ProviderAWS, account2.Provider)

		actualAccount2, err := client.GetAccount(account2.ID)
		require.NoError(t, err)
		require.Equal(t, account2.ID, actualAccount2.ID)
		require.Equal(t, model.ProviderAWS, actualAccount2.Provider)
		require.Equal(t, model.AccountStateCreationRequested, actualAccount2.State)
		require.Equal(t, "service-catalog-id", actualAccount2.ProviderMetadataAWS.ServiceCatalogProductID)
		require.Equal(t, false, actualAccount2.AccountMetadata.Provision)

		time.Sleep(1 * time.Millisecond)

		account3, err := client.CreateAccount(&model.CreateAccountRequest{
			Provider:                model.ProviderAWS,
			ServiceCatalogProductID: "service-catalog-id",
		})
		require.NoError(t, err)
		require.NotNil(t, account3)
		require.Equal(t, model.ProviderAWS, account3.Provider)

		actualAccount3, err := client.GetAccount(account3.ID)
		require.NoError(t, err)
		require.Equal(t, account3.ID, actualAccount3.ID)
		require.Equal(t, model.ProviderAWS, actualAccount3.Provider)
		require.Equal(t, model.AccountStateCreationRequested, actualAccount3.State)
		require.Equal(t, "service-catalog-id", actualAccount3.ProviderMetadataAWS.ServiceCatalogProductID)
		require.Equal(t, false, actualAccount3.AccountMetadata.Provision)

		time.Sleep(1 * time.Millisecond)

		t.Run("get accounts, page 0, perPage 2, exclude deleted", func(t *testing.T) {
			accounts, err := client.GetAccounts(&model.GetAccountsRequest{
				Page:           0,
				PerPage:        2,
				IncludeDeleted: false,
			})
			require.NoError(t, err)
			require.Equal(t, []*model.Account{account1, account2}, accounts)
		})

		t.Run("get accounts, page 1, perPage 2, exclude deleted", func(t *testing.T) {
			accounts, err := client.GetAccounts(&model.GetAccountsRequest{
				Page:           1,
				PerPage:        2,
				IncludeDeleted: false,
			})
			require.NoError(t, err)
			require.Equal(t, []*model.Account{account3}, accounts)
		})

		t.Run("delete account", func(t *testing.T) {
			account2.State = model.AccountStateStable
			err := sqlStore.UpdateAccount(account2)
			require.NoError(t, err)

			err = client.DeleteAccount(account2.ID)
			require.NoError(t, err)

			account2, err = client.GetAccount(account2.ID)
			require.NoError(t, err)
			require.Equal(t, model.AccountStateDeletionRequested, account2.State)
		})

		t.Run("get accounts after deletion request", func(t *testing.T) {
			t.Run("page 0, perPage 2, exclude deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           0,
					PerPage:        2,
					IncludeDeleted: false,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account1, account2}, accounts)
			})

			t.Run("page 1, perPage 2, exclude deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           1,
					PerPage:        2,
					IncludeDeleted: false,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account3}, accounts)
			})

			t.Run("page 0, perPage 2, include deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           0,
					PerPage:        2,
					IncludeDeleted: true,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account1, account2}, accounts)
			})

			t.Run("page 1, perPage 2, include deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           1,
					PerPage:        2,
					IncludeDeleted: true,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account3}, accounts)
			})
		})

		err = sqlStore.DeleteAccount(account2.ID)
		require.NoError(t, err)

		account2, err = client.GetAccount(account2.ID)
		require.NoError(t, err)
		require.NotEqual(t, 0, account2.DeleteAt)

		t.Run("get accounts after actual deletion", func(t *testing.T) {
			t.Run("page 0, perPage 2, exclude deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           0,
					PerPage:        2,
					IncludeDeleted: false,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account1, account3}, accounts)
			})

			t.Run("page 1, perPage 2, exclude deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           1,
					PerPage:        2,
					IncludeDeleted: false,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{}, accounts)
			})

			t.Run("page 0, perPage 2, include deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           0,
					PerPage:        2,
					IncludeDeleted: true,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account1, account2}, accounts)
			})

			t.Run("page 1, perPage 2, include deleted", func(t *testing.T) {
				accounts, err := client.GetAccounts(&model.GetAccountsRequest{
					Page:           1,
					PerPage:        2,
					IncludeDeleted: true,
				})
				require.NoError(t, err)
				require.Equal(t, []*model.Account{account3}, accounts)
			})
		})
	})
}

func TestCreateAccount(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := store.MakeTestSQLStore(t, logger)
	defer store.CloseConnection(t, sqlStore)

	router := mux.NewRouter()
	api.Register(router, &api.Context{
		Store:      sqlStore,
		Supervisor: &mockSupervisor{},
		Logger:     logger,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := model.NewClient(ts.URL)

	t.Run("invalid payload", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/accounts", ts.URL), "application/json", bytes.NewReader([]byte("invalid")))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty payload", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/accounts", ts.URL), "application/json", bytes.NewReader([]byte("")))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid provider", func(t *testing.T) {
		_, err := client.CreateAccount(&model.CreateAccountRequest{
			Provider:                "invalid",
			ServiceCatalogProductID: "service-catalog-id",
			Provision:               false,
		})
		require.EqualError(t, err, "failed with status code 400")
	})

	t.Run("invalid service catalog", func(t *testing.T) {
		_, err := client.CreateAccount(&model.CreateAccountRequest{
			Provider:                model.ProviderAWS,
			ServiceCatalogProductID: "",
			Provision:               false,
		})
		require.EqualError(t, err, "failed with status code 400")
	})

	t.Run("valid", func(t *testing.T) {
		account, err := client.CreateAccount(&model.CreateAccountRequest{
			Provider:                model.ProviderAWS,
			ServiceCatalogProductID: "service-catalog-id",
			Provision:               true,
		})
		require.NoError(t, err)
		require.Equal(t, model.ProviderAWS, account.Provider)
		require.Equal(t, model.AccountStateCreationRequested, account.State)
		require.Equal(t, "service-catalog-id", account.ProviderMetadataAWS.ServiceCatalogProductID)
		require.Equal(t, true, account.AccountMetadata.Provision)
	})
}

func TestRetryCreateAccount(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := store.MakeTestSQLStore(t, logger)
	defer store.CloseConnection(t, sqlStore)

	router := mux.NewRouter()
	api.Register(router, &api.Context{
		Store:      sqlStore,
		Supervisor: &mockSupervisor{},
		Logger:     logger,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := model.NewClient(ts.URL)

	account1, err := client.CreateAccount(&model.CreateAccountRequest{
		Provider:                model.ProviderAWS,
		ServiceCatalogProductID: "service-catalog-id",
		Provision:               false,
	})
	require.NoError(t, err)

	t.Run("unknown account", func(t *testing.T) {
		err := client.RetryCreateAccount(model.NewID())
		require.EqualError(t, err, "failed with status code 404")
	})

	t.Run("while locked", func(t *testing.T) {
		account1.State = model.AccountStateStable
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		lockerID := model.NewID()

		locked, err := sqlStore.LockAccount(account1.ID, lockerID)
		require.NoError(t, err)
		require.True(t, locked)
		defer func() {
			unlocked, err := sqlStore.UnlockAccount(account1.ID, lockerID, false)
			require.NoError(t, err)
			require.True(t, unlocked)
		}()

		err = client.RetryCreateAccount(account1.ID)
		require.EqualError(t, err, "failed with status code 409")
	})

	t.Run("while creating", func(t *testing.T) {
		account1.State = model.AccountStateCreationRequested
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		err = client.RetryCreateAccount(account1.ID)
		require.NoError(t, err)

		account1, err = client.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, model.AccountStateCreationRequested, account1.State)
	})

	t.Run("while stable", func(t *testing.T) {
		account1.State = model.AccountStateStable
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		err = client.RetryCreateAccount(account1.ID)
		require.EqualError(t, err, "failed with status code 400")
	})

	t.Run("while creation failed", func(t *testing.T) {
		account1.State = model.AccountStateCreationFailed
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		err = client.RetryCreateAccount(account1.ID)
		require.NoError(t, err)

		account1, err = client.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, model.AccountStateCreationRequested, account1.State)
	})
}

func TestProvisionAccount(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := store.MakeTestSQLStore(t, logger)
	defer store.CloseConnection(t, sqlStore)

	router := mux.NewRouter()
	api.Register(router, &api.Context{
		Store:      sqlStore,
		Supervisor: &mockSupervisor{},
		Logger:     logger,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := model.NewClient(ts.URL)

	account1, err := client.CreateAccount(&model.CreateAccountRequest{
		Provider:                model.ProviderAWS,
		ServiceCatalogProductID: "service-catalog-id",
		Provision:               false,
		Subnet:                  "10.0.0.0/24",
	})
	require.NoError(t, err)

	t.Run("unknown account", func(t *testing.T) {
		accountResp, err := client.ProvisionAccount(model.NewID(), nil)
		require.EqualError(t, err, "failed with status code 404")
		assert.Nil(t, accountResp)
	})

	t.Run("while locked", func(t *testing.T) {
		account1.State = model.AccountStateStable
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		lockerID := model.NewID()

		locked, err := sqlStore.LockAccount(account1.ID, lockerID)
		require.NoError(t, err)
		require.True(t, locked)
		defer func() {
			unlocked, err := sqlStore.UnlockAccount(account1.ID, lockerID, false)
			require.NoError(t, err)
			require.True(t, unlocked)
		}()

		accountResp, err := client.ProvisionAccount(account1.ID, nil)
		require.EqualError(t, err, "failed with status code 409")
		assert.Nil(t, accountResp)
	})

	t.Run("while api-security-locked", func(t *testing.T) {
		err = sqlStore.LockAccountAPI(account1.ID)
		require.NoError(t, err)

		accountResp, err := client.ProvisionAccount(account1.ID, nil)
		require.EqualError(t, err, "failed with status code 403")
		assert.Nil(t, accountResp)

		err = sqlStore.UnlockAccountAPI(account1.ID)
		require.NoError(t, err)
	})

	t.Run("while provisioning", func(t *testing.T) {
		account1.State = model.AccountStateProvisioningRequested
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		accountResp, err := client.ProvisionAccount(account1.ID, nil)
		require.NoError(t, err)
		assert.NotNil(t, accountResp)

		account1, err = client.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, model.AccountStateProvisioningRequested, account1.State)
	})

	t.Run("after provisioning failed", func(t *testing.T) {
		account1.State = model.AccountStateProvisioningFailed
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		accountResp, err := client.ProvisionAccount(account1.ID, nil)
		require.NoError(t, err)
		assert.NotNil(t, accountResp)

		account1, err = client.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, model.AccountStateProvisioningRequested, account1.State)
	})

	t.Run("while stable", func(t *testing.T) {
		account1.State = model.AccountStateStable
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		accountResp, err := client.ProvisionAccount(account1.ID, nil)
		require.NoError(t, err)
		assert.NotNil(t, accountResp)

		account1, err = client.GetAccount(account1.ID)
		require.NoError(t, err)
		require.Equal(t, model.AccountStateProvisioningRequested, account1.State)
	})

	t.Run("while deleting", func(t *testing.T) {
		account1.State = model.AccountStateDeletionRequested
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		accountResp, err := client.ProvisionAccount(account1.ID, nil)
		require.EqualError(t, err, "failed with status code 400")
		assert.Nil(t, accountResp)
	})
}

func TestDeleteCluster(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := store.MakeTestSQLStore(t, logger)
	defer store.CloseConnection(t, sqlStore)

	router := mux.NewRouter()
	api.Register(router, &api.Context{
		Store:      sqlStore,
		Supervisor: &mockSupervisor{},
		Logger:     logger,
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := model.NewClient(ts.URL)

	account1, err := client.CreateAccount(&model.CreateAccountRequest{
		Provider:                model.ProviderAWS,
		ServiceCatalogProductID: "service-catalog-id",
		Provision:               false,
	})
	require.NoError(t, err)

	t.Run("unknown account", func(t *testing.T) {
		err := client.DeleteAccount(model.NewID())
		require.EqualError(t, err, "failed with status code 404")
	})

	t.Run("while locked", func(t *testing.T) {
		account1.State = model.AccountStateStable
		err = sqlStore.UpdateAccount(account1)
		require.NoError(t, err)

		lockerID := model.NewID()

		locked, err := sqlStore.LockAccount(account1.ID, lockerID)
		require.NoError(t, err)
		require.True(t, locked)
		defer func() {
			unlocked, err := sqlStore.UnlockAccount(account1.ID, lockerID, false)
			require.NoError(t, err)
			require.True(t, unlocked)

			account1, err = client.GetAccount(account1.ID)
			require.NoError(t, err)
			require.Equal(t, int64(0), account1.LockAcquiredAt)
		}()

		err = client.DeleteAccount(account1.ID)
		require.EqualError(t, err, "failed with status code 409")
	})

	t.Run("while api-security-locked", func(t *testing.T) {
		err = sqlStore.LockAccountAPI(account1.ID)
		require.NoError(t, err)

		err := client.DeleteAccount(account1.ID)
		require.EqualError(t, err, "failed with status code 403")

		err = sqlStore.UnlockAccountAPI(account1.ID)
		require.NoError(t, err)
	})

	// valid unlocked states
	states := []string{
		model.AccountStateStable,
		model.AccountStateCreationRequested,
		model.AccountStateCreationFailed,
		model.AccountStateProvisioningFailed,
		model.AccountStateDeletionRequested,
		model.AccountStateDeletionFailed,
	}

	t.Run("from a valid, unlocked state", func(t *testing.T) {
		for _, state := range states {
			t.Run(state, func(t *testing.T) {
				account1.State = state
				err = sqlStore.UpdateAccount(account1)
				require.NoError(t, err)

				err = client.DeleteAccount(account1.ID)
				require.NoError(t, err)

				account1, err = client.GetAccount(account1.ID)
				require.NoError(t, err)
				require.Equal(t, model.AccountStateDeletionRequested, account1.State)
			})
		}
	})
}
