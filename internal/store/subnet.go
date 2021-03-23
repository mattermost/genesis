// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"database/sql"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
)

var subnetSelect sq.SelectBuilder

func init() {
	subnetSelect = sq.
		Select("SubnetPool.ID", "CIDR", "Used", "ParentSubnet", "SubnetMetadataRaw",
			"LockAcquiredBy", "LockAcquiredAt").
		From("SubnetPool")
}

// RawSubnetMetadata is the raw byte metadata for a subnet.
type RawSubnetMetadata struct {
	SubnetMetadataRaw []byte
}

type rawSubnet struct {
	*model.Subnet
	*RawSubnetMetadata
}
type rawSubnets []*rawSubnet

func buildSubnetRawMetadata(subnet *model.Subnet) (*RawSubnetMetadata, error) {
	subnetMetadataJSON, err := json.Marshal(subnet.SubnetMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal SubnetMetadata")
	}

	return &RawSubnetMetadata{
		SubnetMetadataRaw: subnetMetadataJSON,
	}, nil
}

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

// GetSubnet fetches the given subnet by subnet range.
func (sqlStore *SQLStore) GetSubnet(cidr string) (*model.Subnet, error) {
	var rawSubnet rawSubnet
	err := sqlStore.getBuilder(sqlStore.db, &rawSubnet, subnetSelect.Where("ID = ?", cidr))
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to get subnet by range")
	}

	return rawSubnet.toSubnet()
}

// GetSubnets fetches the given page of added subnets. The first page is 0.
func (sqlStore *SQLStore) GetSubnets(filter *model.SubnetFilter) ([]*model.Subnet, error) {
	builder := subnetSelect.
		OrderBy("CreateAt ASC")
	builder = sqlStore.applySubnetsFilter(builder, filter)

	var rawSubnets rawSubnets
	err := sqlStore.selectBuilder(sqlStore.db, &rawSubnets, builder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for subnets")
	}

	if filter.Free {
		builder = builder.Where("Used = false")
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

// AddSubnet records the given subnet to the database.
func (sqlStore *SQLStore) AddSubnet(subnet *model.Subnet) error {
	tx, err := sqlStore.beginTransaction(sqlStore.db)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.RollbackUnlessCommitted()

	err = sqlStore.addSubnet(tx, subnet)
	if err != nil {
		return errors.Wrap(err, "failed to add subnet")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit the transaction")
	}

	return nil
}

// addSubnet records the given subnet to the database.
func (sqlStore *SQLStore) addSubnet(execer execer, subnet *model.Subnet) error {
	subnet.CreateAt = GetMillis()
	subnet.ID = model.NewID()

	rawMetadata, err := buildSubnetRawMetadata(subnet)
	if err != nil {
		return errors.Wrap(err, "unable to build raw cluster metadata")
	}

	_, err = sqlStore.execBuilder(execer, sq.
		Insert("SubnetPool").
		SetMap(map[string]interface{}{
			"ID":                subnet.ID,
			"CIDR":              subnet.CIDR,
			"Used":              subnet.Used,
			"ParentSubnet":      subnet.ParentSubnet,
			"SubnetMetadataRaw": rawMetadata.SubnetMetadataRaw,
			"CreateAt":          subnet.CreateAt,
			"LockAcquiredBy":    nil,
			"LockAcquiredAt":    0,
		}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create subnet")
	}

	return nil
}

// UpdateSubnet updates the given subnet in the database.
func (sqlStore *SQLStore) UpdateSubnet(subnet *model.Subnet) error {
	rawMetadata, err := buildSubnetRawMetadata(subnet)
	if err != nil {
		return errors.Wrap(err, "unable to build raw cluster metadata")
	}

	_, err = sqlStore.execBuilder(sqlStore.db, sq.
		Update("SubnetPool").
		SetMap(map[string]interface{}{
			"Used":              subnet.Used,
			"SubnetMetadataRaw": rawMetadata.SubnetMetadataRaw,
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
