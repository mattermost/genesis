// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package genesis

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/mattermost/genesis/model"
)

// SplitParentSubnet splits a parent subnet into usable provisioning VPCs.
func SplitParentSubnet(parentSubnet *model.ParentSubnet) ([]model.Subnet, error) {
	logger := logger.WithField("parent-subnet", parentSubnet.ID)
	logger.Info(parentSubnet.ID)
	logger.Infof("Splitting parent subnet %s", parentSubnet.CIDR)
	_, base, err := net.ParseCIDR(parentSubnet.CIDR)
	if err != nil {
		return nil, err
	}

	parentRange := strings.Split(parentSubnet.CIDR, "/")
	intParentRange, err := strconv.Atoi(parentRange[1])
	if err != nil {
		return nil, err
	}

	subs, err := splitSubnet(base, parentSubnet.SplitRange-intParentRange, logger)
	if err != nil {
		return nil, err
	}

	var subnets []model.Subnet
	for _, sub := range subs {
		subnet := model.Subnet{
			CIDR:         fmt.Sprintf("%s/%d", &sub.IP, parentSubnet.SplitRange),
			ParentSubnet: parentSubnet.CIDR,
			CreateAt:     parentSubnet.CreateAt,
		}
		subnets = append(subnets, subnet)

	}
	logger.Infof("Finished splitting parent subnet %s", parentSubnet.CIDR)

	return subnets, nil
}
