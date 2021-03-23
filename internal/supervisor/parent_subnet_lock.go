// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package supervisor

import (
	log "github.com/sirupsen/logrus"
)

type parentSubnetLockStore interface {
	LockParentSubnet(subnet, lockerID string) (bool, error)
	UnlockParentSubnet(subnet, lockerID string, force bool) (bool, error)
}

type parentSubnetLock struct {
	subnet   string
	lockerID string
	store    parentSubnetLockStore
	logger   log.FieldLogger
}

func newParentSubnetLock(subnet, lockerID string, store parentSubnetStore, logger log.FieldLogger) *parentSubnetLock {
	return &parentSubnetLock{
		subnet:   subnet,
		lockerID: lockerID,
		store:    store,
		logger:   logger,
	}
}

func (l *parentSubnetLock) TryLock() bool {
	locked, err := l.store.LockParentSubnet(l.subnet, l.lockerID)
	if err != nil {
		l.logger.WithError(err).Error("failed to lock parent subnet")
		return false
	}

	return locked
}

func (l *parentSubnetLock) Unlock() {
	unlocked, err := l.store.UnlockParentSubnet(l.subnet, l.lockerID, false)
	if err != nil {
		l.logger.WithError(err).Error("failed to unlock parent subnet")
	} else if unlocked != true {
		l.logger.Error("failed to release lock for parent subnet")
	}
}
