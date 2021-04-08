// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
)

var subnetSelect sq.SelectBuilder

func init() {
	subnetSelect = sq.
		Select("SubnetPool.ID", "CIDR", "AccountID", "ParentSubnet",
			"LockAcquiredBy", "LockAcquiredAt").
		From("SubnetPool")
}

type rawSubnet struct {
	*model.Subnet
}
type rawSubnets []*rawSubnet

func (r *rawSubnet) toSubnet() (*model.Subnet, error) {
	return r.Subnet, nil
}

func (rc *rawSubnets) toSubnets() ([]*model.Subnet, error) {
	var subnets []*model.Subnet
	for _, rawSubnet := range *rc {
		subnet, err := rawSubnet.toSubnet()
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

// GetSubnet fetches a random available subnet.
func (sqlStore *SQLStore) GetRandomAvailableSubnet() (*model.Subnet, error) {
	filter := &model.SubnetFilter{
		Page:    1,
		PerPage: 1,
		Free:    true,
	}

	subnet, err := sqlStore.GetSubnets(filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for free subnet")
	}

	return subnet[0], nil
}

// GetSubnet fetches a subnet by its provided CIDR range.
func (sqlStore *SQLStore) GetSubnetByCIDR(cidr string) (*model.Subnet, error) {
	var rawSubnet rawSubnet
	err := sqlStore.getBuilder(sqlStore.db, &rawSubnet, subnetSelect.Where("CIDR = ?", cidr))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get subnet by range")
	}

	return rawSubnet.toSubnet()
}

// GetSubnet fetches the given subnet by subnet range.
func (sqlStore *SQLStore) GetSubnet(id string) (*model.Subnet, error) {
	var rawSubnet rawSubnet
	err := sqlStore.getBuilder(sqlStore.db, &rawSubnet, subnetSelect.Where("ID = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get subnet by id")
	}

	return rawSubnet.toSubnet()
}

// GetSubnets fetches the given page of added subnets. The first page is 0.
func (sqlStore *SQLStore) GetSubnets(filter *model.SubnetFilter) ([]*model.Subnet, error) {
	builder := subnetSelect.
		OrderBy("CreateAt ASC")
	builder = sqlStore.applySubnetsFilter(builder, filter)

	if filter.Free {
		builder = builder.Where("AccountID == '' ")
	}

	var rawSubnets rawSubnets
	err := sqlStore.selectBuilder(sqlStore.db, &rawSubnets, builder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for subnets")
	}

	return rawSubnets.toSubnets()
}

func (sqlStore *SQLStore) applySubnetsFilter(builder sq.SelectBuilder, filter *model.SubnetFilter) sq.SelectBuilder {
	if filter.PerPage != model.AllPerPage {
		builder = builder.
			Limit(uint64(filter.PerPage)).
			Offset(uint64(filter.Page * filter.PerPage))
	}

	return builder
}

// addSubnet records the given subnet to the database.
func (sqlStore *SQLStore) addSubnets(execer execer, subnets *[]model.Subnet) error {
	for _, subnet := range *subnets {
		subnet.CreateAt = GetMillis()
		subnet.ID = model.NewID()

		_, err := sqlStore.execBuilder(execer, sq.
			Insert("SubnetPool").
			SetMap(map[string]interface{}{
				"ID":             subnet.ID,
				"CIDR":           subnet.CIDR,
				"AccountID":      subnet.AccountID,
				"ParentSubnet":   subnet.ParentSubnet,
				"CreateAt":       subnet.CreateAt,
				"LockAcquiredBy": nil,
				"LockAcquiredAt": 0,
			}),
		)
		if err != nil {
			return errors.Wrap(err, "failed to add subnet")
		}
	}

	return nil
}

// UpdateSubnet updates the given subnet in the database.
func (sqlStore *SQLStore) UpdateSubnet(subnet *model.Subnet) error {
	_, err := sqlStore.execBuilder(sqlStore.db, sq.
		Update("SubnetPool").
		SetMap(map[string]interface{}{
			"AccountID": subnet.AccountID,
		}).
		Where("ID = ?", subnet.ID),
	)
	if err != nil {
		return errors.Wrap(err, "failed to update subnet")
	}

	return nil
}

// LockSubnet marks the subnet as locked for exclusive use by the caller.
func (sqlStore *SQLStore) LockSubnet(subnet, lockerID string) (bool, error) {
	return sqlStore.lockRows("SubnetPool", []string{subnet}, lockerID)
}

// UnlockSubnet releases a lock previously acquired against a caller.
func (sqlStore *SQLStore) UnlockSubnet(subnet, lockerID string, force bool) (bool, error) {
	return sqlStore.unlockRows("SubnetPool", []string{subnet}, lockerID, force)
}
