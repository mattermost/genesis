// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package supervisor

import (
	log "github.com/sirupsen/logrus"
)

type accountLockStore interface {
	LockAccount(accountID, lockerID string) (bool, error)
	UnlockAccount(accountID, lockerID string, force bool) (bool, error)
}

type accountLock struct {
	accountID string
	lockerID  string
	store     accountLockStore
	logger    log.FieldLogger
}

func newAccountLock(accountID, lockerID string, store accountLockStore, logger log.FieldLogger) *accountLock {
	return &accountLock{
		accountID: accountID,
		lockerID:  lockerID,
		store:     store,
		logger:    logger,
	}
}

func (l *accountLock) TryLock() bool {
	locked, err := l.store.LockAccount(l.accountID, l.lockerID)
	if err != nil {
		l.logger.WithError(err).Error("failed to lock account")
		return false
	}

	return locked
}

func (l *accountLock) Unlock() {
	unlocked, err := l.store.UnlockAccount(l.accountID, l.lockerID, false)
	if err != nil {
		l.logger.WithError(err).Error("failed to unlock account")
	} else if !unlocked {
		l.logger.Error("failed to release lock for account")
	}
}
