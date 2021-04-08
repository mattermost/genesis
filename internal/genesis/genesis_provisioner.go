package genesis

import (
	model "github.com/mattermost/genesis/model"
	log "github.com/sirupsen/logrus"
)

// GenProvisioner provisions accounts using AWS API and Terraform.
type GenProvisioner struct {
	accountCreation  model.AccountCreation
	accountProvision model.AccountProvision
	logger           log.FieldLogger
}

// NewGenesisProvisioner creates a new GenProvisioner.
func NewGenesisProvisioner(accountCreation model.AccountCreation, accountProvision model.AccountProvision, logger log.FieldLogger) *GenProvisioner {
	logger = logger.WithField("provisioner", "genesis")

	return &GenProvisioner{
		accountCreation:  accountCreation,
		accountProvision: accountProvision,
		logger:           logger,
	}
}
