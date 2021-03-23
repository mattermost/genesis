// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package supervisor

import (
	"time"

	"github.com/mattermost/genesis/internal/aws"
	"github.com/mattermost/genesis/internal/webhook"

	"github.com/mattermost/genesis/model"
	log "github.com/sirupsen/logrus"
)

// accountStore abstracts the database operations required to query accounts.
type accountStore interface {
	GetAccount(accountID string) (*model.Account, error)
	GetUnlockedAccountsPendingWork() ([]*model.Account, error)
	GetAccounts(accountFilter *model.AccountFilter) ([]*model.Account, error)
	UpdateAccount(account *model.Account) error
	LockAccount(accountID, lockerID string) (bool, error)
	UnlockAccount(accountID string, lockerID string, force bool) (bool, error)
	DeleteAccount(accountID string) error

	GetWebhooks(filter *model.WebhookFilter) ([]*model.Webhook, error)
}

// accountProvisioner abstracts the provisioning operations required by the account supervisor.
type accountProvisioner interface {
	PrepareAccount(account *model.Account) bool
	CreateAccount(account *model.Account, aws aws.AWS) error
	ProvisionAccount(account *model.Account, aws aws.AWS) error
	DeleteAccount(account *model.Account, aws aws.AWS) error
}

// AccountSupervisor finds accounts pending work and effects the required changes.
//
// The degree of parallelism is controlled by a weighted semaphore, intended to be shared with
// other clients needing to coordinate background jobs.
type AccountSupervisor struct {
	store       accountStore
	provisioner accountProvisioner
	aws         aws.AWS
	instanceID  string
	logger      log.FieldLogger
}

// NewAccountSupervisor creates a new AccountSupervisor.
func NewAccountSupervisor(store accountStore, accountProvisioner accountProvisioner, aws aws.AWS, instanceID string, logger log.FieldLogger) *AccountSupervisor {
	return &AccountSupervisor{
		store:       store,
		provisioner: accountProvisioner,
		aws:         aws,
		instanceID:  instanceID,
		logger:      logger,
	}
}

// Shutdown performs graceful shutdown tasks for the account supervisor.
func (s *AccountSupervisor) Shutdown() {
	s.logger.Debug("Shutting down account supervisor")
}

// Do looks for work to be done on any pending accounts and attempts to schedule the required work.
func (s *AccountSupervisor) Do() error {
	accounts, err := s.store.GetUnlockedAccountsPendingWork()
	if err != nil {
		s.logger.WithError(err).Warn("Failed to query for accounts pending work")
		return nil
	}

	for _, account := range accounts {
		s.Supervise(account)
	}

	return nil
}

// Supervise schedules the required work on the given account.
func (s *AccountSupervisor) Supervise(account *model.Account) {
	logger := s.logger.WithFields(log.Fields{
		"account": account.ID,
	})

	lock := newAccountLock(account.ID, s.instanceID, s.store, logger)
	if !lock.TryLock() {
		return
	}
	defer lock.Unlock()

	// Before working on the account, it is crucial that we ensure that it was
	// not updated to a new state by another genesis server.
	originalState := account.State
	account, err := s.store.GetAccount(account.ID)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get refreshed account")
		return
	}
	if account.State != originalState {
		logger.WithField("oldAccountState", originalState).
			WithField("newAccountState", account.State).
			Warn("Another provisioner has worked on this account; skipping...")
		return
	}

	logger.Debugf("Supervising account in state %s", account.State)

	newState := s.transitionAccount(account, logger)

	account, err = s.store.GetAccount(account.ID)
	if err != nil {
		logger.WithError(err).Warnf("failed to get account and thus persist state %s", newState)
		return
	}

	if account.State == newState {
		return
	}

	oldState := account.State
	account.State = newState
	err = s.store.UpdateAccount(account)
	if err != nil {
		logger.WithError(err).Warnf("failed to set account state to %s", newState)
		return
	}

	environment, err := s.aws.GetCloudEnvironmentName()
	if err != nil {
		logger.WithError(err).Error("getting the AWS Cloud environment")
		return
	}

	webhookPayload := &model.WebhookPayload{
		Type:      model.TypeAccount,
		ID:        account.ID,
		NewState:  newState,
		OldState:  oldState,
		Timestamp: time.Now().UnixNano(),
		ExtraData: map[string]string{"Environment": environment},
	}
	err = webhook.SendToAllWebhooks(s.store, webhookPayload, logger.WithField("webhookEvent", webhookPayload.NewState))
	if err != nil {
		logger.WithError(err).Error("Unable to process and send webhooks")
	}

	logger.Debugf("Transitioned account from %s to %s", oldState, newState)
}

// Do works with the given account to transition it to a final state.
func (s *AccountSupervisor) transitionAccount(account *model.Account, logger log.FieldLogger) string {
	switch account.State {
	case model.AccountStateCreationRequested:
		return s.createAccount(account, logger)
	case model.AccountStateProvisioningRequested:
		return s.provisionAccount(account, logger)
	case model.AccountStateDeletionRequested:
		return s.deleteAccount(account, logger)
	case model.AccountStateRefreshMetadata:
		return s.refreshAccountMetadata(account, logger)
	default:
		logger.Warnf("Found account pending work in unexpected state %s", account.State)
		return account.State
	}
}

func (s *AccountSupervisor) createAccount(account *model.Account, logger log.FieldLogger) string {
	var err error

	if s.provisioner.PrepareAccount(account) {
		err = s.store.UpdateAccount(account)
		if err != nil {
			logger.WithError(err).Error("Failed to record updated account after creation")
			return model.AccountStateCreationFailed
		}
	}

	err = s.provisioner.CreateAccount(account, s.aws)
	if err != nil {
		logger.WithError(err).Error("Failed to create account")
		return model.AccountStateCreationFailed
	}

	logger.Info("Finished creating account")
	if account.AccountMetadata.Provision {
		return s.provisionAccount(account, logger)
	}
	return model.AccountStateStable
}

func (s *AccountSupervisor) provisionAccount(account *model.Account, logger log.FieldLogger) string {
	err := s.provisioner.ProvisionAccount(account, s.aws)
	if err != nil {
		logger.WithError(err).Error("Failed to provision account")
		return model.AccountStateProvisioningFailed
	}

	logger.Info("Finished provisioning account")
	return s.refreshAccountMetadata(account, logger)
}

func (s *AccountSupervisor) refreshAccountMetadata(account *model.Account, logger log.FieldLogger) string {
	err := s.store.UpdateAccount(account)
	if err != nil {
		logger.WithError(err).Error("Failed to save updated account metadata")
		return model.AccountStateProvisioningFailed
	}

	return model.AccountStateStable
}

func (s *AccountSupervisor) deleteAccount(account *model.Account, logger log.FieldLogger) string {
	err := s.provisioner.DeleteAccount(account, s.aws)
	if err != nil {
		logger.WithError(err).Error("Failed to delete account")
		return model.AccountStateDeletionFailed
	}

	err = s.store.DeleteAccount(account.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to record updated account after deletion")
		return model.AccountStateDeletionFailed
	}

	logger.Info("Finished deleting account")
	return model.AccountStateDeleted
}
