// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package genesis

import (
	"net"
	"strconv"
	"strings"

	"github.com/mattermost/genesis/model"
)

// AddParentSubnet adds a parent subnet.
func (provisioner *GenesisProvisioner) AddParentSubnet(subnet *model.ParentSubnet) error {
	logger := provisioner.logger.WithField("subnet", subnet)
	logger.Infof("Adding subnet %s", subnet.ID)

	return nil
}

// SplitParentSubnet splits a parent subnet into usable provisioning VPCs.
func (provisioner *GenesisProvisioner) SplitParentSubnet(subnet *model.ParentSubnet) ([]net.IPNet, error) {
	logger := provisioner.logger.WithField("subnet", subnet)
	logger.Infof("Splitting subnet %s", subnet.CIDR)
	_, base, err := net.ParseCIDR(subnet.CIDR)
	if err != nil {
		return nil, err
	}

	parentRange := strings.Split(subnet.CIDR, "/")
	intParentRange, err := strconv.Atoi(parentRange[1])
	if err != nil {
		return nil, err
	}

	subnets, err := splitSubnet(base, subnet.SplitRange-intParentRange, logger)
	if err != nil {
		return nil, err
	}

	return subnets, nil
}
