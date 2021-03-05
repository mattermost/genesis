// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package store

import (
	"testing"

	"github.com/mattermost/genesis/internal/testlib"
	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountDTOs(t *testing.T) {
	logger := testlib.MakeLogger(t)
	sqlStore := MakeTestSQLStore(t, logger)
	defer CloseConnection(t, sqlStore)

	t.Run("get unknown account DTO", func(t *testing.T) {
		account, err := sqlStore.GetAccountDTO("unknown")
		require.NoError(t, err)
		require.Nil(t, account)
	})

	account1 := &model.Account{
		Provider:            "aws",
		Provisioner:         "genesis",
		ProviderMetadataAWS: &model.AWSMetadata{ServiceCatalogProductID: "prod-12345"},
		AccountMetadata:     &model.AccountMetadata{Provision: true},
		State:               model.AccountStateCreationRequested,
	}

	err := sqlStore.CreateAccount(account1)
	require.NoError(t, err)

	t.Run("get account DTO", func(t *testing.T) {
		accountDTO, err := sqlStore.GetAccountDTO(account1.ID)
		require.NoError(t, err)
		assert.Equal(t, account1, accountDTO.Account)
	})

	t.Run("get account DTOs", func(t *testing.T) {
		accountDTOs, err := sqlStore.GetAccountDTOs(&model.AccountFilter{PerPage: model.AllPerPage, IncludeDeleted: true})
		require.NoError(t, err)
		assert.Equal(t, 1, len(accountDTOs))
	})
}
