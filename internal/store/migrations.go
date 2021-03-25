// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"github.com/blang/semver"
)

type migration struct {
	fromVersion   semver.Version
	toVersion     semver.Version
	migrationFunc func(execer) error
}

// migrations defines the set of migrations necessary to advance the database to the latest
// expected version.
//
// Note that the canonical schema is currently obtained by applying all migrations to an empty
// database.
var migrations = []migration{
	{semver.MustParse("0.0.0"), semver.MustParse("0.1.0"), func(e execer) error {
		_, err := e.Exec(`
			CREATE TABLE System (
				Key VARCHAR(64) PRIMARY KEY,
				Value VARCHAR(1024) NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE Account (
				ID CHAR(26) PRIMARY KEY,
				State TEXT NOT NULL,
				Provider TEXT NOT NULL,
				ProviderMetadataRaw BYTEA NULL,
				Provisioner TEXT NOT NULL,
				AccountMetadataRaw BYTEA NULL,
				CreateAt BIGINT NOT NULL,
				DeleteAt BIGINT NOT NULL,
				APISecurityLock BOOLEAN NOT NULL,
				LockAcquiredBy CHAR(26) NULL,
				LockAcquiredAt BIGINT NOT NULL
			);
		`)
		if err != nil {
			return err
		}

		// Add webhook table.
		_, err = e.Exec(`
			CREATE TABLE Webhooks (
				ID TEXT PRIMARY KEY,
				OwnerID TEXT NOT NULL,
				URL TEXT NOT NULL,
				CreateAt BIGINT NOT NULL,
				DeleteAt BIGINT NOT NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE UNIQUE INDEX Webhook_URL_DeleteAt ON Webhooks (URL, DeleteAt);
		`)
		if err != nil {
			return err
		}

		return nil
	}},
	{semver.MustParse("0.1.0"), semver.MustParse("0.2.0"), func(e execer) error {
		_, err := e.Exec(`
			CREATE TABLE ParentSubnet (
				ID TEXT PRIMARY KEY,
				CIDR TEXT NOT NULL,
				SplitRange INT NOT NULL,
				CreateAt BIGINT NOT NULL,
				LockAcquiredBy CHAR(26) NULL,
				LockAcquiredAt BIGINT NOT NULL
			);
		`)
		if err != nil {
			return err
		}

		_, err = e.Exec(`
			CREATE TABLE SubnetPool (
				ID TEXT PRIMARY KEY,
				CIDR TEXT NOT NULL,
				AccountID TEXT NOT NULL,
				VPCID TEXT NOT NULL,
				ParentSubnet TEXT NOT NULL,
				CreateAt BIGINT NOT NULL,
				LockAcquiredBy CHAR(26) NULL,
				LockAcquiredAt BIGINT NOT NULL
			);
	`)
		if err != nil {
			return err
		}

		return nil
	}},
}
