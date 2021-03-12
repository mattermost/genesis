// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/genesis/internal/webhook"
	"github.com/mattermost/genesis/model"
)

// initAccount registers account endpoints on the given router.
func initAccount(apiRouter *mux.Router, context *Context) {
	addContext := func(handler contextHandlerFunc) *contextHandler {
		return newContextHandler(context, handler)
	}

	accountsRouter := apiRouter.PathPrefix("/accounts").Subrouter()
	accountsRouter.Handle("", addContext(handleGetAccounts)).Methods("GET")
	accountsRouter.Handle("", addContext(handleCreateAccount)).Methods("POST")

	accountRouter := apiRouter.PathPrefix("/account/{account:[A-Za-z0-9]{26}}").Subrouter()
	accountRouter.Handle("", addContext(handleGetAccount)).Methods("GET")
	accountRouter.Handle("", addContext(handleRetryCreateAccount)).Methods("POST")
	accountRouter.Handle("/provision", addContext(handleProvisionAccount)).Methods("POST")

	accountRouter.Handle("", addContext(handleDeleteAccount)).Methods("DELETE")
}

// handleGetAccount responds to GET /api/account/{account}, returning the account in question.
func handleGetAccount(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["account"]
	c.Logger = c.Logger.WithField("account", accountID)

	account, err := c.Store.GetAccount(accountID)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query account")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if account == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	outputJSON(c, w, account)
}

// handleGetAccounts responds to GET /api/accounts, returning the specified page of accounts.
func handleGetAccounts(c *Context, w http.ResponseWriter, r *http.Request) {
	page, perPage, includeDeleted, _, err := parsePaging(r.URL)
	if err != nil {
		c.Logger.WithError(err).Error("failed to parse paging parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := &model.AccountFilter{
		Page:           page,
		PerPage:        perPage,
		IncludeDeleted: includeDeleted,
	}

	accounts, err := c.Store.GetAccounts(filter)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query accounts")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if accounts == nil {
		accounts = []*model.Account{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	outputJSON(c, w, accounts)
}

// handleCreateAccount responds to POST /api/accounts, beginning the process of creating a new
// account.
// sample body:
// {
//		"provider": "aws",
// }
func handleCreateAccount(c *Context, w http.ResponseWriter, r *http.Request) {
	createAccountRequest, err := model.NewCreateAccountRequestFromReader(r.Body)
	if err != nil {
		c.Logger.WithError(err).Error("failed to decode request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	account := model.Account{
		Provider: createAccountRequest.Provider,
		ProviderMetadataAWS: &model.AWSMetadata{
			ServiceCatalogProductID: createAccountRequest.ServiceCatalogProductID,
			AWSAccountID:            "",
			AccountProductID:        "",
		},
		AccountMetadata: &model.AccountMetadata{
			Provision: createAccountRequest.Provision,
		},
		Provisioner:     "genesis",
		APISecurityLock: createAccountRequest.APISecurityLock,
		State:           model.AccountStateCreationRequested,
	}

	err = c.Store.CreateAccount(&account)
	if err != nil {
		c.Logger.WithError(err).Error("failed to create account")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	webhookPayload := &model.WebhookPayload{
		Type:      model.TypeAccount,
		ID:        account.ID,
		NewState:  model.AccountStateCreationRequested,
		OldState:  "n/a",
		Timestamp: time.Now().UnixNano(),
		ExtraData: map[string]string{"Environment": c.Environment},
	}
	err = webhook.SendToAllWebhooks(c.Store, webhookPayload, c.Logger.WithField("webhookEvent", webhookPayload.NewState))
	if err != nil {
		c.Logger.WithError(err).Error("Unable to process and send webhooks")
	}

	c.Supervisor.Do()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	outputJSON(c, w, account)
}

// handleRetryCreateAccount responds to POST /api/account/{account}, retrying a previously
// failed creation.
func handleRetryCreateAccount(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["account"]
	c.Logger = c.Logger.WithField("account", accountID)

	account, status, unlockOnce := lockAccount(c, accountID)
	if status != 0 {
		w.WriteHeader(status)
		return
	}
	defer unlockOnce()

	newState := model.AccountStateCreationRequested

	if !account.ValidTransitionState(newState) {
		c.Logger.Warnf("unable to retry account creation while in state %s", account.State)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if account.State != newState {
		webhookPayload := &model.WebhookPayload{
			Type:      model.TypeAccount,
			ID:        account.ID,
			NewState:  newState,
			OldState:  account.State,
			Timestamp: time.Now().UnixNano(),
			ExtraData: map[string]string{"Environment": c.Environment},
		}
		account.State = newState

		err := c.Store.UpdateAccount(account)
		if err != nil {
			c.Logger.WithError(err).Errorf("failed to retry account creation")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = webhook.SendToAllWebhooks(c.Store, webhookPayload, c.Logger.WithField("webhookEvent", webhookPayload.NewState))
		if err != nil {
			c.Logger.WithError(err).Error("Unable to process and send webhooks")
		}
	}

	// Notify even if we didn't make changes, to expedite even the no-op operations above.
	unlockOnce()
	c.Supervisor.Do()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	outputJSON(c, w, account)
}

// handleProvisionCluster responds to POST /api/account/{account}/provision,
// provisioning AWS resources on a previously-created account.
func handleProvisionAccount(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["account"]
	c.Logger = c.Logger.WithField("account", accountID)

	account, status, unlockOnce := lockAccount(c, accountID)
	if status != 0 {
		w.WriteHeader(status)
		return
	}
	defer unlockOnce()

	if account.APISecurityLock {
		logSecurityLockConflict("account", c.Logger)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, err := model.NewProvisionAccountRequestFromReader(r.Body)
	if err != nil {
		c.Logger.WithError(err).Error("failed to deserialize account provision request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newState := model.AccountStateProvisioningRequested

	if !account.ValidTransitionState(newState) {
		c.Logger.Warnf("unable to provision account while in state %s", account.State)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if account.State != newState {
		webhookPayload := &model.WebhookPayload{
			Type:      model.TypeAccount,
			ID:        account.ID,
			NewState:  newState,
			OldState:  account.State,
			Timestamp: time.Now().UnixNano(),
			ExtraData: map[string]string{"Environment": c.Environment},
		}
		account.State = newState
		account.AccountMetadata.Provision = true

		err := c.Store.UpdateAccount(account)
		if err != nil {
			c.Logger.WithError(err).Errorf("failed to mark account provisioning state")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = webhook.SendToAllWebhooks(c.Store, webhookPayload, c.Logger.WithField("webhookEvent", webhookPayload.NewState))
		if err != nil {
			c.Logger.WithError(err).Error("Unable to process and send webhooks")
		}
	}

	// Notify even if we didn't make changes, to expedite even the no-op operations above.
	unlockOnce()
	c.Supervisor.Do()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	outputJSON(c, w, account)
}

// handleDeleteAccount responds to DELETE /api/account/{account}, beginning the process of
// deleting the account.
func handleDeleteAccount(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID := vars["account"]
	c.Logger = c.Logger.WithField("account", accountID)

	account, status, unlockOnce := lockAccount(c, accountID)
	if status != 0 {
		w.WriteHeader(status)
		return
	}
	defer unlockOnce()

	if account.APISecurityLock {
		logSecurityLockConflict("account", c.Logger)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	newState := model.AccountStateDeletionRequested

	if !account.ValidTransitionState(newState) {
		c.Logger.Warnf("unable to delete account while in state %s", account.State)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: will be used soon
	// genesisResources, err := c.Store.GetGenesisResources(&model.GenesisResourcesFilter{
	// 	AccountID:      account.ID,
	// 	IncludeDeleted: false,
	// 	PerPage:        model.AllPerPage,
	// })
	// if err != nil {
	// 	c.Logger.WithError(err).Error("failed to get genesis resources")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// if len(genesisResources) != 0 {
	// 	c.Logger.Errorf("unable to delete account while it still has %d genesis resources", len(genesisResources))
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }

	if account.State != newState {
		webhookPayload := &model.WebhookPayload{
			Type:      model.TypeAccount,
			ID:        account.ID,
			NewState:  newState,
			OldState:  account.State,
			Timestamp: time.Now().UnixNano(),
			ExtraData: map[string]string{"Environment": c.Environment},
		}
		account.State = newState

		err := c.Store.UpdateAccount(account)
		if err != nil {
			c.Logger.WithError(err).Error("failed to mark account for deletion")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = webhook.SendToAllWebhooks(c.Store, webhookPayload, c.Logger.WithField("webhookEvent", webhookPayload.NewState))
		if err != nil {
			c.Logger.WithError(err).Error("Unable to process and send webhooks")
		}
	}

	unlockOnce()
	c.Supervisor.Do()

	w.WriteHeader(http.StatusAccepted)
}
