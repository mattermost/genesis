// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package supervisor_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/mattermost/genesis/internal/aws"
	"github.com/mattermost/genesis/internal/store"
	"github.com/mattermost/genesis/internal/supervisor"
	"github.com/mattermost/genesis/internal/testlib"
	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/require"
)

type mockAccountStore struct {
	Account                     *model.Account
	UnlockedAccountsPendingWork []*model.Account
	Accounts                    []*model.Account

	UnlockChan         chan interface{}
	UpdateAccountCalls int
}

type mockAWS struct{}

func (a *mockAWS) AssumeRole(roleArn string) (*credentials.Credentials, error) {
	return nil, nil
}

func (a *mockAWS) GetAccountAliases() (*iam.ListAccountAliasesOutput, error) {
	return nil, nil
}

func (a *mockAWS) GetAccountID() (string, error) {
	return "", nil
}

func (a *mockAWS) AssociateTGWShare(resourceShareARN, principalID string) error {
	return nil
}
func (a *mockAWS) DisassociateTGWShare(resourceShareARN, principalID string) error {
	return nil
}

func (a *mockAWS) GetCloudEnvironmentName() (string, error) {
	return "", nil
}

func (s *mockAccountStore) GetAccount(AccountID string) (*model.Account, error) {
	return s.Account, nil
}

func (s *mockAccountStore) GetUnlockedAccountsPendingWork() ([]*model.Account, error) {
	return s.UnlockedAccountsPendingWork, nil
}

func (s *mockAccountStore) GetAccounts(AccountFilter *model.AccountFilter) ([]*model.Account, error) {
	return s.Accounts, nil
}

func (s *mockAccountStore) UpdateAccount(Account *model.Account) error {
	s.UpdateAccountCalls++
	return nil
}

func (s *mockAccountStore) LockAccount(AccountID, lockerID string) (bool, error) {
	return true, nil
}

func (s *mockAccountStore) UnlockAccount(AccountID string, lockerID string, force bool) (bool, error) {
	if s.UnlockChan != nil {
		close(s.UnlockChan)
	}
	return true, nil
}

func (s *mockAccountStore) DeleteAccount(AccountID string) error {
	return nil
}

func (s *mockAccountStore) GetRandomAvailableSubnet() (*model.Subnet, error) {
	return nil, nil
}

func (s *mockAccountStore) GetSubnetByCIDR(cidr string) (*model.Subnet, error) {
	return nil, nil
}

func (s *mockAccountStore) UpdateSubnet(subnet *model.Subnet) error {
	return nil
}

func (s *mockAccountStore) GetWebhooks(filter *model.WebhookFilter) ([]*model.Webhook, error) {
	return nil, nil
}

type mockAccountProvisioner struct{}

func (p *mockAccountProvisioner) PrepareAccount(Account *model.Account) bool {
	return true
}

func (p *mockAccountProvisioner) CreateAccount(Account *model.Account, aws aws.AWS) error {
	return nil
}

func (p *mockAccountProvisioner) ProvisionAccount(Account *model.Account, aws aws.AWS) error {
	return nil
}

func (p *mockAccountProvisioner) DeleteAccount(Account *model.Account, aws aws.AWS) error {
	return nil
}

func TestAccountSupervisorDo(t *testing.T) {
	t.Run("no Accounts pending work", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		mockStore := &mockAccountStore{}

		supervisor := supervisor.NewAccountSupervisor(mockStore, &mockAccountProvisioner{}, &mockAWS{}, "instanceID", logger)
		err := supervisor.Do()
		require.NoError(t, err)

		require.Equal(t, 0, mockStore.UpdateAccountCalls)
	})

	t.Run("mock Account creation", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		mockStore := &mockAccountStore{}

		mockStore.UnlockedAccountsPendingWork = []*model.Account{{
			ID:              model.NewID(),
			State:           model.AccountStateCreationRequested,
			AccountMetadata: &model.AccountMetadata{},
		}}
		mockStore.Account = mockStore.UnlockedAccountsPendingWork[0]
		mockStore.UnlockChan = make(chan interface{})

		supervisor := supervisor.NewAccountSupervisor(mockStore, &mockAccountProvisioner{}, &mockAWS{}, "instanceID", logger)
		err := supervisor.Do()
		require.NoError(t, err)

		<-mockStore.UnlockChan
		require.Equal(t, 3, mockStore.UpdateAccountCalls)
	})
	t.Run("mock Account creation and provision", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		mockStore := &mockAccountStore{}

		mockStore.UnlockedAccountsPendingWork = []*model.Account{{
			ID:    model.NewID(),
			State: model.AccountStateCreationRequested,
			AccountMetadata: &model.AccountMetadata{
				Provision: true,
				Subnet:    "10.0.0.0/24",
			},
		}}
		mockStore.Account = mockStore.UnlockedAccountsPendingWork[0]
		mockStore.UnlockChan = make(chan interface{})

		supervisor := supervisor.NewAccountSupervisor(mockStore, &mockAccountProvisioner{}, &mockAWS{}, "instanceID", logger)
		err := supervisor.Do()
		require.NoError(t, err)

		<-mockStore.UnlockChan
		require.Equal(t, 3, mockStore.UpdateAccountCalls)
	})
}

func TestAccountSupervisorSupervise(t *testing.T) {
	testCases := []struct {
		Description   string
		InitialState  string
		ExpectedState string
	}{
		{"unexpected state", model.AccountStateStable, model.AccountStateStable},
		{"creation requested", model.AccountStateCreationRequested, model.AccountStateStable},
		{"provision requested", model.AccountStateProvisioningRequested, model.AccountStateStable},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			logger := testlib.MakeLogger(t)
			sqlStore := store.MakeTestSQLStore(t, logger)
			supervisor := supervisor.NewAccountSupervisor(sqlStore, &mockAccountProvisioner{}, &mockAWS{}, "instanceID", logger)

			Account := &model.Account{
				Provider:        model.ProviderAWS,
				State:           tc.InitialState,
				AccountMetadata: &model.AccountMetadata{},
			}
			err := sqlStore.CreateAccount(Account)
			require.NoError(t, err)

			supervisor.Supervise(Account)

			Account, err = sqlStore.GetAccount(Account.ID)
			require.NoError(t, err)
			require.Equal(t, tc.ExpectedState, Account.State)
		})
	}

	t.Run("state has changed since Account was selected to be worked on", func(t *testing.T) {
		logger := testlib.MakeLogger(t)
		sqlStore := store.MakeTestSQLStore(t, logger)
		supervisor := supervisor.NewAccountSupervisor(sqlStore, &mockAccountProvisioner{}, &mockAWS{}, "instanceID", logger)

		Account := &model.Account{
			Provider: model.ProviderAWS,
			State:    model.AccountStateDeletionRequested,
		}
		err := sqlStore.CreateAccount(Account)
		require.NoError(t, err)

		// The stored Account is AccountStateDeletionRequested, so we will pass
		// in a Account with state of AccountStateCreationRequested to simulate
		// stale state.
		Account.State = model.AccountStateCreationRequested

		supervisor.Supervise(Account)

		Account, err = sqlStore.GetAccount(Account.ID)
		require.NoError(t, err)
		require.Equal(t, model.AccountStateDeletionRequested, Account.State)
	})
}
