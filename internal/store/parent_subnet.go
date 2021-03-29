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

var parentSubnetSelect sq.SelectBuilder

func init() {
	parentSubnetSelect = sq.
		Select("ParentSubnet.ID", "CIDR", "SplitRange", "CreateAt",
			"LockAcquiredBy", "LockAcquiredAt").
		From("ParentSubnet")
}

type rawParentSubnet struct {
	*model.ParentSubnet
}
type rawParentSubnets []*rawParentSubnet

func (r *rawParentSubnet) toParentSubnet() (model.ParentSubnet, error) {
	return *r.ParentSubnet, nil
}

func (rc *rawParentSubnets) toParentSubnets() ([]model.ParentSubnet, error) {
	var parentSubnets []model.ParentSubnet
	for _, rawParentSubnet := range *rc {
		parentSubnet, err := rawParentSubnet.toParentSubnet()
		if err != nil {
			return nil, err
		}
		parentSubnets = append(parentSubnets, parentSubnet)
	}

	return parentSubnets, nil
}

// GetParentSubnet fetches the given parent subnet by subnet range.
func (sqlStore *SQLStore) GetParentSubnet(id string) (model.ParentSubnet, error) {
	var rawParentSubnet rawParentSubnet
	err := sqlStore.getBuilder(sqlStore.db, &rawParentSubnet, parentSubnetSelect.Where("ID = ?", id))
	if err == sql.ErrNoRows {
		return model.ParentSubnet{}, nil
	} else if err != nil {
		return model.ParentSubnet{}, errors.Wrap(err, "failed to get parent subnet by range")
	}

	return rawParentSubnet.toParentSubnet()
}

// GetParentSubnets fetches the given page of added parent subnets. The first page is 0.
func (sqlStore *SQLStore) GetParentSubnets(filter *model.ParentSubnetFilter) ([]model.ParentSubnet, error) {
	builder := parentSubnetSelect.
		OrderBy("CreateAt ASC")
	builder = sqlStore.applyParentSubnetsFilter(builder, filter)

	var rawParentSubnets rawParentSubnets
	err := sqlStore.selectBuilder(sqlStore.db, &rawParentSubnets, builder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for parent subnets")
	}

	return rawParentSubnets.toParentSubnets()
}

func (sqlStore *SQLStore) applyParentSubnetsFilter(builder sq.SelectBuilder, filter *model.ParentSubnetFilter) sq.SelectBuilder {
	if filter.PerPage != model.AllPerPage {
		builder = builder.
			Limit(uint64(filter.PerPage)).
			Offset(uint64(filter.Page * filter.PerPage))
	}

	return builder
}

// AddParentSubnet records the given parent subnet to the database.
func (sqlStore *SQLStore) AddParentSubnet(parentSubnet *model.ParentSubnet, subnets *[]model.Subnet) error {
	tx, err := sqlStore.beginTransaction(sqlStore.db)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.RollbackUnlessCommitted()

	if err = sqlStore.addParentSubnet(tx, parentSubnet); err != nil {
		return errors.Wrap(err, "failed to add parent subnet")
	}

	if err = sqlStore.addSubnets(tx, subnets); err != nil {
		return errors.Wrap(err, "failed to add subnets")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit the transaction")
	}

	return nil
}

// addParentSubnet records the given parent subnet to the database.
func (sqlStore *SQLStore) addParentSubnet(execer execer, parentSubnet *model.ParentSubnet) error {
	parentSubnet.CreateAt = GetMillis()

	if _, err := sqlStore.execBuilder(execer, sq.
		Insert("ParentSubnet").
		SetMap(map[string]interface{}{
			"ID":             parentSubnet.ID,
			"CIDR":           parentSubnet.CIDR,
			"SplitRange":     parentSubnet.SplitRange,
			"CreateAt":       parentSubnet.CreateAt,
			"LockAcquiredBy": nil,
			"LockAcquiredAt": 0,
		}),
	); err != nil {
		return errors.Wrap(err, "failed to create parent subnet")
	}

	return nil
}

// LockParentSubnet marks the parent subnet as locked for exclusive use by the caller.
func (sqlStore *SQLStore) LockParentSubnet(subnet, lockerID string) (bool, error) {
	return sqlStore.lockRows("ParentSubnet", []string{subnet}, lockerID)
}

// UnlockParentSubnet releases a lock previously acquired against a caller.
func (sqlStore *SQLStore) UnlockParentSubnet(subnet, lockerID string, force bool) (bool, error) {
	return sqlStore.unlockRows("ParentSubnet", []string{subnet}, lockerID, force)
}
