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

// lockParentSubnet synchronizes access to the given parent subnet across potentially
// multiple genesis servers.
func lockParentSubnet(c *Context, parentSubnet string) (model.ParentSubnet, int, func()) {
	parentSub, err := c.Store.GetParentSubnet(parentSubnet)
	if err != nil {
		c.Logger.WithError(err).Error("failed to query parent subnet")
		return model.ParentSubnet{}, http.StatusInternalServerError, nil
	}
	if &parentSub == nil {
		return model.ParentSubnet{}, http.StatusNotFound, nil
	}

	locked, err := c.Store.LockParentSubnet(parentSubnet, c.RequestID)
	if err != nil {
		c.Logger.WithError(err).Error("failed to lock parent subnet")
		return model.ParentSubnet{}, http.StatusInternalServerError, nil
	} else if !locked {
		c.Logger.Error("failed to acquire lock for parent subnet")
		return model.ParentSubnet{}, http.StatusConflict, nil
	}

	unlockOnce := sync.Once{}

	return parentSub, 0, func() {
		unlockOnce.Do(func() {
			unlocked, err := c.Store.UnlockParentSubnet(parentSub.ID, c.RequestID, false)
			if err != nil {
				c.Logger.WithError(err).Errorf("failed to unlock parent subnet")
			} else if unlocked != true {
				c.Logger.Error("failed to release lock for parent subnet")
			}
		})
	}
}
