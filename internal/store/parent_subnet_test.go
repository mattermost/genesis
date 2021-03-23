// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"testing"

	"github.com/mattermost/genesis/internal/testlib"
	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/require"
)

func TestParentSubnets(t *testing.T) {
	t.Run("get unknown parent subnet", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		parentSubnet, err := sqlStore.GetParentSubnet("unknown")
		require.NoError(t, err)
		require.Nil(t, parentSubnet)
	})

	t.Run("get parent subnets", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 8,
			State:      model.ParentSubnetStateAdditionRequested,
		}

		err := sqlStore.AddParentSubnet(parentSubnet1)
		require.NoError(t, err)

		actualParentSubnet1, err := sqlStore.GetParentSubnet(parentSubnet1.ID)
		require.NoError(t, err)
		require.Equal(t, parentSubnet1, actualParentSubnet1)

		actualParentSubnets, err := sqlStore.GetParentSubnets(&model.ParentSubnetFilter{Page: 0, PerPage: 0})
		require.NoError(t, err)
		require.Empty(t, actualParentSubnets)

		actualParentSubnets, err = sqlStore.GetParentSubnets(&model.ParentSubnetFilter{Page: 0, PerPage: 1})
		require.NoError(t, err)
		require.Equal(t, []*model.ParentSubnet{parentSubnet1}, actualParentSubnets)

		actualParentSubnets, err = sqlStore.GetParentSubnets(&model.ParentSubnetFilter{Page: 0, PerPage: 10})
		require.NoError(t, err)
		require.Equal(t, []*model.ParentSubnet{parentSubnet1}, actualParentSubnets)

		actualParentSubnets, err = sqlStore.GetParentSubnets(&model.ParentSubnetFilter{PerPage: model.AllPerPage})
		require.NoError(t, err)
		require.Equal(t, []*model.ParentSubnet{parentSubnet1}, actualParentSubnets)
	})

	t.Run("update parent subnets", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := MakeTestSQLStore(t, logger)
		defer CloseConnection(t, sqlStore)

		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 8,
			State:      model.ParentSubnetStateAdditionRequested,
		}

		err := sqlStore.AddParentSubnet(parentSubnet1)
		require.NoError(t, err)

		parentSubnet1.State = model.ParentSubnetStateSplitRequested

		err = sqlStore.UpdateParentSubnet(parentSubnet1)
		require.NoError(t, err)

		actualParentSubnet1, err := sqlStore.GetParentSubnet(parentSubnet1.ID)
		require.NoError(t, err)
		require.Equal(t, parentSubnet1, actualParentSubnet1)
	})
}

func TestLockParentCidr(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := MakeTestSQLStore(t, logger)

	lockerID1 := model.NewID()
	lockerID2 := model.NewID()

	parentSubnet1 := &model.ParentSubnet{}
	err := sqlStore.AddParentSubnet(parentSubnet1)
	require.NoError(t, err)

	parentSubnet2 := &model.ParentSubnet{}
	err = sqlStore.AddParentSubnet(parentSubnet2)
	require.NoError(t, err)

	t.Run("parent subnets should start unlocked", func(t *testing.T) {
		parentSubnet1, err = sqlStore.GetParentSubnet(parentSubnet1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), parentSubnet1.LockAcquiredAt)
		require.Nil(t, parentSubnet1.LockAcquiredBy)

		parentSubnet2, err = sqlStore.GetParentSubnet(parentSubnet2.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), parentSubnet2.LockAcquiredAt)
		require.Nil(t, parentSubnet2.LockAcquiredBy)
	})

	t.Run("lock an unlocked parent subnet", func(t *testing.T) {
		locked, err := sqlStore.LockParentSubnet(parentSubnet1.ID, lockerID1)
		require.NoError(t, err)
		require.True(t, locked)

		parentSubnet1, err = sqlStore.GetParentSubnet(parentSubnet1.ID)
		require.NoError(t, err)
		require.NotEqual(t, int64(0), parentSubnet1.LockAcquiredAt)
		require.Equal(t, lockerID1, *parentSubnet1.LockAcquiredBy)
	})

	t.Run("lock a previously locked parent subnet", func(t *testing.T) {
		t.Run("by the same locker", func(t *testing.T) {
			locked, err := sqlStore.LockParentSubnet(parentSubnet1.ID, lockerID1)
			require.NoError(t, err)
			require.False(t, locked)
		})

		t.Run("by a different locker", func(t *testing.T) {
			locked, err := sqlStore.LockParentSubnet(parentSubnet1.ID, lockerID2)
			require.NoError(t, err)
			require.False(t, locked)
		})
	})

	t.Run("lock a second parent subnet from a different locker", func(t *testing.T) {
		locked, err := sqlStore.LockParentSubnet(parentSubnet2.ID, lockerID2)
		require.NoError(t, err)
		require.True(t, locked)

		parentSubnet2, err = sqlStore.GetParentSubnet(parentSubnet2.ID)
		require.NoError(t, err)
		require.NotEqual(t, int64(0), parentSubnet2.LockAcquiredAt)
		require.Equal(t, lockerID2, *parentSubnet2.LockAcquiredBy)
	})

	t.Run("unlock the first parent subnet", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockParentSubnet(parentSubnet1.ID, lockerID1, false)
		require.NoError(t, err)
		require.True(t, unlocked)

		parentSubnet1, err = sqlStore.GetParentSubnet(parentSubnet1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), parentSubnet1.LockAcquiredAt)
		require.Nil(t, parentSubnet1.LockAcquiredBy)
	})

	t.Run("unlock the first parent subnet again", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockParentSubnet(parentSubnet2.ID, lockerID1, false)
		require.NoError(t, err)
		require.False(t, unlocked)

		parentSubnet2, err = sqlStore.GetParentSubnet(parentSubnet2.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), parentSubnet2.LockAcquiredAt)
		require.Nil(t, parentSubnet2.LockAcquiredBy)
	})

	t.Run("force unlock the first parent subnet again", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockParentSubnet(parentSubnet1.ID, lockerID1, true)
		require.NoError(t, err)
		require.False(t, unlocked)

		parentSubnet1, err = sqlStore.GetParentSubnet(parentSubnet1.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), parentSubnet1.LockAcquiredAt)
		require.Nil(t, parentSubnet1.LockAcquiredBy)
	})

	t.Run("unlock the second parent subnet from the wrong locker", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockParentSubnet(parentSubnet2.ID, lockerID1, false)
		require.NoError(t, err)
		require.False(t, unlocked)

		parentSubnet2, err = sqlStore.GetParentSubnet(parentSubnet2.ID)
		require.NoError(t, err)
		require.NotEqual(t, int64(0), parentSubnet2.LockAcquiredAt)
		require.Equal(t, lockerID2, *parentSubnet2.LockAcquiredBy)
	})

	t.Run("force unlock the second parent subnet from the wrong locker", func(t *testing.T) {
		unlocked, err := sqlStore.UnlockParentSubnet(parentSubnet2.ID, lockerID1, true)
		require.NoError(t, err)
		require.True(t, unlocked)

		parentSubnet2, err = sqlStore.GetParentSubnet(parentSubnet2.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), parentSubnet2.LockAcquiredAt)
		require.Nil(t, parentSubnet2.LockAcquiredBy)
	})
}
