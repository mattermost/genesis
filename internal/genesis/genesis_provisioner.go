package genesis

import (
	log "github.com/sirupsen/logrus"
)

// GenProvisioner provisions accounts using AWS API and Terraform.
type GenProvisioner struct {
	ssoUserEmail          string
	ssoFirstName          string
	ssoLastName           string
	managedOU             string
	controlTowerRole      string
	controlTowerAccountID string
	resourceShareID       string
	coreAccountID         string
	logger                log.FieldLogger
}

// NewGenesisProvisioner creates a new GenProvisioner.
func NewGenesisProvisioner(ssoUserEmail, ssoFirstName, ssoLastName, managedOU, controlTowerRole, controlTowerAccountID, resourceShareID, coreAccountID string, logger log.FieldLogger) *GenProvisioner {
	logger = logger.WithField("provisioner", "genesis")

	return &GenProvisioner{
		ssoUserEmail:          ssoUserEmail,
		ssoFirstName:          ssoFirstName,
		ssoLastName:           ssoLastName,
		managedOU:             managedOU,
		controlTowerRole:      controlTowerRole,
		controlTowerAccountID: controlTowerAccountID,
		resourceShareID:       resourceShareID,
		coreAccountID:         coreAccountID,
		logger:                logger,
	}
}
