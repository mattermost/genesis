// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	// "github.com/mattermost/mattermost-cloud/k8s"

	"github.com/mattermost/genesis/model"
	"github.com/sirupsen/logrus"
)

// Supervisor describes the interface to notify the background jobs of an actionable change.
type Supervisor interface {
	Do() error
}

// Store describes the interface required to persist changes made via API requests.
type Store interface {
	CreateAccount(account *model.Account) error
	GetAccount(accountID string) (*model.Account, error)
	GetAccountDTO(accountID string) (*model.AccountDTO, error)
	GetAccounts(filter *model.AccountFilter) ([]*model.Account, error)
	GetAccountDTOs(filter *model.AccountFilter) ([]*model.AccountDTO, error)
	UpdateAccount(account *model.Account) error
	LockAccount(accountID, lockerID string) (bool, error)
	UnlockAccount(accountID, lockerID string, force bool) (bool, error)
	LockAccountAPI(accountID string) error
	UnlockAccountAPI(accountID string) error
	DeleteAccount(accountID string) error

	CreateWebhook(webhook *model.Webhook) error
	GetWebhook(webhookID string) (*model.Webhook, error)
	GetWebhooks(filter *model.WebhookFilter) ([]*model.Webhook, error)
	DeleteWebhook(webhookID string) error
}

// TODO: will be used

// Genesis describes the interface required to communicate with the AWS account.
type Genesis interface {
}

// Context provides the API with all necessary data and interfaces for responding to requests.
//
// It is cloned before each request, allowing per-request changes such as logger annotations.
type Context struct {
	Store       Store
	Supervisor  Supervisor
	Genesis     Genesis
	RequestID   string
	Environment string
	Logger      logrus.FieldLogger
}

// Clone creates a shallow copy of context, allowing clones to apply per-request changes.
func (c *Context) Clone() *Context {
	return &Context{
		Store:      c.Store,
		Supervisor: c.Supervisor,
		Genesis:    c.Genesis,
		Logger:     c.Logger,
	}
}
