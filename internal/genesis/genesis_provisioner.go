package genesis

import (
	log "github.com/sirupsen/logrus"
)

// GenesisProvisioner provisions accounts using AWS API and Terraform.
type GenesisProvisioner struct {
	ssoUserEmail          string
	ssoFirstName          string
	ssoLastName           string
	managedOU             string
	controlTowerRole      string
	controlTowerAccountID string
	logger                log.FieldLogger
}

// NewGenesisProvisioner creates a new GenesisProvisioner.
func NewGenesisProvisioner(ssoUserEmail, ssoFirstName, ssoLastName, managedOU, controlTowerRole, controlTowerAccountID string, logger log.FieldLogger) *GenesisProvisioner {
	logger = logger.WithField("provisioner", "genesis")

	return &GenesisProvisioner{
		ssoUserEmail:          ssoUserEmail,
		ssoFirstName:          ssoFirstName,
		ssoLastName:           ssoLastName,
		managedOU:             managedOU,
		controlTowerRole:      controlTowerRole,
		controlTowerAccountID: controlTowerAccountID,
		logger:                logger,
	}
}
