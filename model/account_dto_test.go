// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccountDTOFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		accountDTO, err := AccountDTOFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, &AccountDTO{}, accountDTO)
	})

	t.Run("invalid request", func(t *testing.T) {
		accountDTO, err := AccountDTOFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, accountDTO)
	})

	t.Run("request", func(t *testing.T) {
		accountDTO, err := AccountDTOFromReader(bytes.NewReader([]byte(
			`{"ID":"id","Provider":"aws"}`,
		)))
		require.NoError(t, err)
		require.Equal(t, &AccountDTO{Account: &Account{ID: "id", Provider: "aws"}}, accountDTO)
	})
}

func TestAccountsDTOFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		accountDTOs, err := AccountDTOsFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, []*AccountDTO{}, accountDTOs)
	})

	t.Run("invalid request", func(t *testing.T) {
		accountDTOs, err := AccountDTOsFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, accountDTOs)
	})

	t.Run("request", func(t *testing.T) {
		accountDTOs, err := AccountDTOsFromReader(bytes.NewReader([]byte(
			`[{"ID":"id1","Provider":"aws"}, {"ID":"id2","Provider":"aws"}]`,
		)))
		require.NoError(t, err)
		require.Equal(t, []*AccountDTO{
			{
				Account: &Account{ID: "id1", Provider: "aws"},
			},
			{
				Account: &Account{ID: "id2", Provider: "aws"},
			},
		}, accountDTOs)
	})
}
