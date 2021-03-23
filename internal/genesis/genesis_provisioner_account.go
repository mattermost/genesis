// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package genesis

import (
	"github.com/mattermost/genesis/internal/aws"
	"github.com/mattermost/genesis/model"
)

// TODO: will be used soon
// PrepareAccount ensures an account object is ready for provisioning.
func (provisioner *GenesisProvisioner) PrepareAccount(account *model.Account) bool {
	return true
}

// CreateAccount creates an account using AWS API and terraform.
func (provisioner *GenesisProvisioner) CreateAccount(account *model.Account, awsClient aws.AWS) error {
	logger := provisioner.logger.WithField("account", account.ID)
	err := createAccount(provisioner, account, logger, awsClient)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAccount deletes an account using AWS API and terraform.
func (provisioner *GenesisProvisioner) DeleteAccount(account *model.Account, awsClient aws.AWS) error {
	logger := provisioner.logger.WithField("account", account.ID)
	err := deleteAccount(provisioner, account, logger, awsClient)
	if err != nil {
		return err
	}
	return nil
}

// ProvisionAccount deletes an account using AWS API and terraform.
func (provisioner *GenesisProvisioner) ProvisionAccount(account *model.Account, awsClient aws.AWS) error {
	logger := provisioner.logger.WithField("account", account.ID)
	logger.Infof("Provisioning account %s", account.ID)
	return nil
}
