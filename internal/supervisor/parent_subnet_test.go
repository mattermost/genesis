// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package supervisor_test

import (
	"net"
	"testing"

	"github.com/mattermost/genesis/internal/store"
	"github.com/mattermost/genesis/internal/supervisor"
	"github.com/mattermost/genesis/internal/testlib"
	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/require"
)

type mockParentSubnetStore struct {
	ParentSubnet                     *model.ParentSubnet
	UnlockedParentSubnetsPendingWork []*model.ParentSubnet
	ParentSubnets                    []*model.ParentSubnet

	UnlockChan              chan interface{}
	UpdateParentSubnetCalls int
}

func (s *mockParentSubnetStore) GetParentSubnet(ParentSubnetID string) (*model.ParentSubnet, error) {
	return s.ParentSubnet, nil
}

func (s *mockParentSubnetStore) GetUnlockedParentSubnetsPendingWork() ([]*model.ParentSubnet, error) {
	return s.UnlockedParentSubnetsPendingWork, nil
}

func (s *mockParentSubnetStore) GetParentSubnets(ParentSubnetFilter *model.ParentSubnetFilter) ([]*model.ParentSubnet, error) {
	return s.ParentSubnets, nil
}

func (s *mockParentSubnetStore) UpdateParentSubnet(ParentSubnet *model.ParentSubnet) error {
	s.UpdateParentSubnetCalls++
	return nil
}

func (s *mockParentSubnetStore) LockParentSubnet(ParentSubnetID, lockerID string) (bool, error) {
	return true, nil
}

func (s *mockParentSubnetStore) UnlockParentSubnet(ParentSubnetID string, lockerID string, force bool) (bool, error) {
	if s.UnlockChan != nil {
		close(s.UnlockChan)
	}
	return true, nil
}

func (s *mockParentSubnetStore) GetWebhooks(filter *model.WebhookFilter) ([]*model.Webhook, error) {
	return nil, nil
}

func (s *mockParentSubnetStore) AddSubnet(subnet *model.Subnet) error {
	return nil
}

type mockParentSubnetProvisioner struct{}

func (p *mockParentSubnetProvisioner) AddParentSubnet(parentSubnet *model.ParentSubnet) error {
	return nil
}

func (p *mockParentSubnetProvisioner) SplitParentSubnet(parentSubnet *model.ParentSubnet) ([]net.IPNet, error) {
	return nil, nil
}

func TestParentSupervisorDo(t *testing.T) {
	t.Run("no parent subnets pending work", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		mockStore := &mockParentSubnetStore{}

		supervisor := supervisor.NewParentSubnetSupervisor(mockStore, &mockParentSubnetProvisioner{}, "instanceID", logger)
		err := supervisor.Do()
		require.NoError(t, err)

		require.Equal(t, 0, mockStore.UpdateParentSubnetCalls)
	})

	t.Run("mock parent subnet addition", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		mockStore := &mockParentSubnetStore{}

		mockStore.UnlockedParentSubnetsPendingWork = []*model.ParentSubnet{{
			ID:    model.NewID(),
			State: model.ParentSubnetStateAdditionRequested,
		}}
		mockStore.ParentSubnet = mockStore.UnlockedParentSubnetsPendingWork[0]
		mockStore.UnlockChan = make(chan interface{})

		supervisor := supervisor.NewParentSubnetSupervisor(mockStore, &mockParentSubnetProvisioner{}, "instanceID", logger)
		err := supervisor.Do()
		require.NoError(t, err)

		<-mockStore.UnlockChan
		require.Equal(t, 1, mockStore.UpdateParentSubnetCalls)
	})
}

func TestParentSubnetSupervisorSupervise(t *testing.T) {
	testCases := []struct {
		Description   string
		InitialState  string
		ExpectedState string
	}{
		{"unexpected state", model.ParentSubnetStateAdded, model.ParentSubnetStateAdded},
		{"creation requested", model.ParentSubnetStateAdditionRequested, model.ParentSubnetStateAdded},
		{"split requested", model.ParentSubnetStateSplitRequested, model.ParentSubnetStateSplitRequested},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			logger := testlib.MakeLogger(t)
			sqlStore := store.MakeTestSQLStore(t, logger)
			supervisor := supervisor.NewParentSubnetSupervisor(sqlStore, &mockParentSubnetProvisioner{}, "instanceID", logger)

			ParentSubnet := &model.ParentSubnet{
				CIDR:  "10.0.0.0/8",
				State: tc.InitialState,
			}
			err := sqlStore.AddParentSubnet(ParentSubnet)
			require.NoError(t, err)

			supervisor.Supervise(ParentSubnet)

			ParentSubnet, err = sqlStore.GetParentSubnet(ParentSubnet.ID)
			require.NoError(t, err)
			require.Equal(t, tc.ExpectedState, ParentSubnet.State)
		})
	}

	t.Run("state has changed since ParentSubnet was selected to be worked on", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := store.MakeTestSQLStore(t, logger)
		supervisor := supervisor.NewParentSubnetSupervisor(sqlStore, &mockParentSubnetProvisioner{}, "instanceID", logger)

		ParentSubnet := &model.ParentSubnet{
			CIDR:  "10.0.0.0/8",
			State: model.ParentSubnetStateAdditionFailed,
		}
		err := sqlStore.AddParentSubnet(ParentSubnet)
		require.NoError(t, err)

		// The stored ParentSubnet is ParentSubnetStateAdditionFailed, so we will pass
		// in a ParentSubnet with state of ParentSubnetStateAdditionRequested to simulate
		// stale state.
		ParentSubnet.State = model.ParentSubnetStateAdditionRequested

		supervisor.Supervise(ParentSubnet)

		ParentSubnet, err = sqlStore.GetParentSubnet(ParentSubnet.ID)
		require.NoError(t, err)
		require.Equal(t, model.ParentSubnetStateAdditionFailed, ParentSubnet.State)
	})
}
