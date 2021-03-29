// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// initSecurity registers security endpoints on the given router.
func initSecurity(apiRouter *mux.Router, context *Context) {
	addContext := func(handler contextHandlerFunc) *contextHandler {
		return newContextHandler(context, handler)
	}

	securityRouter := apiRouter.PathPrefix("/security").Subrouter()

	securityClusterRouter := securityRouter.PathPrefix("/account/{account:[A-Za-z0-9]{26}}").Subrouter()
	securityClusterRouter.Handle("/api/lock", addContext(handleAccountLockAPI)).Methods("POST")
	securityClusterRouter.Handle("/api/unlock", addContext(handleAccountUnlockAPI)).Methods("POST")
}

// handleAccountLockAPI responds to POST /api/security/account/{account}/api/lock,
// locking API changes for this account.
func handleAccountLockAPI(c *Context, w http.ResponseWriter, r *http.Request) {
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

	if !account.APISecurityLock {
		if err := c.Store.LockAccountAPI(account.ID); err != nil {
			c.Logger.WithError(err).Error("failed to lock account API")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// handleAccountUnlockAPI responds to POST /api/security/account/{account}/api/unlock,
// unlocking API changes for this account.
func handleAccountUnlockAPI(c *Context, w http.ResponseWriter, r *http.Request) {
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

	if account.APISecurityLock {
		if err = c.Store.UnlockAccountAPI(account.ID); err != nil {
			c.Logger.WithError(err).Error("failed to unlock account API")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
