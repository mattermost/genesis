// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountClone(t *testing.T) {
	account := &Account{
		Provider:            "aws",
		Provisioner:         "kops",
		ProviderMetadataAWS: &AWSMetadata{ServiceCatalogProductID: "prod-12345"},
	}

	clone := account.Clone()
	require.Equal(t, account, clone)

	// Verify changing pointers in the clone doesn't affect the original.
	clone.ProviderMetadataAWS = &AWSMetadata{ServiceCatalogProductID: "prod-123456"}
	require.NotEqual(t, account, clone)
}

func TestAccountFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		account, err := AccountFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, &Account{}, account)
	})

	t.Run("invalid request", func(t *testing.T) {
		account, err := AccountFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, account)
	})

	t.Run("request", func(t *testing.T) {
		account, err := AccountFromReader(bytes.NewReader([]byte(
			`{"ID":"id","Provider":"aws"}`,
		)))
		require.NoError(t, err)
		require.Equal(t, &Account{ID: "id", Provider: "aws"}, account)
	})
}

func TestAccountsFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		accounts, err := AccountsFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, []*Account{}, accounts)
	})

	t.Run("invalid request", func(t *testing.T) {
		accounts, err := AccountsFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, accounts)
	})

	t.Run("request", func(t *testing.T) {
		account, err := AccountsFromReader(bytes.NewReader([]byte(
			`[{"ID":"id1", "Provider":"aws"}, {"ID":"id2", "Provider":"aws"}]`,
		)))
		require.NoError(t, err)
		require.Equal(t, []*Account{
			{ID: "id1", Provider: "aws"},
			{ID: "id2", Provider: "aws"},
		}, account)
	})
}
