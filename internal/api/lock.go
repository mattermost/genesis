// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	"net/http"
	"sync"

	"github.com/mattermost/genesis/model"
)

// lockAccount synchronizes access to the given account across potentially
// multiple genesis servers.
func lockAccount(c *Context, accountID string) (*model.Account, int, func()) {
	account, err := c.Store.GetAccount(accountID)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query account")
		return nil, http.StatusInternalServerError, nil
	}
	if account == nil {
		return nil, http.StatusNotFound, nil
	}

	locked, err := c.Store.LockAccount(accountID, c.RequestID)
	if err != nil {
		c.Logger.WithError(err).Error("failed to lock account")
		return nil, http.StatusInternalServerError, nil
	} else if !locked {
		c.Logger.Error("failed to acquire lock for account")
		return nil, http.StatusConflict, nil
	}

	unlockOnce := sync.Once{}

	return account, 0, func() {
		unlockOnce.Do(func() {
			unlocked, err := c.Store.UnlockAccount(account.ID, c.RequestID, false)
			if err != nil {
				c.Logger.WithError(err).Errorf("failed to unlock account")
			} else if unlocked != true {
				c.Logger.Error("failed to release lock for account")
			}
		})
	}
}
