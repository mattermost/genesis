// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParentSubnetClone(t *testing.T) {
	parentSubnet := &ParentSubnet{
		CIDR:       "10.0.0.0/8",
		SplitRange: 24,
	}

	clone := parentSubnet.Clone()
	require.Equal(t, parentSubnet, clone)
}

func TestParentSubnetFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		parentSubnet, err := ParentSubnetFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, &ParentSubnet{}, parentSubnet)
	})

	t.Run("invalid request", func(t *testing.T) {
		parentSubnet, err := ParentSubnetFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, parentSubnet)
	})

	t.Run("request", func(t *testing.T) {
		parentSubnet, err := ParentSubnetFromReader(bytes.NewReader([]byte(
			`{"ID":"id"}`,
		)))
		require.NoError(t, err)
		require.Equal(t, &ParentSubnet{ID: "id"}, parentSubnet)
	})
}

func TestParentSubnetsFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		parentSubnets, err := ParentSubnetsFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, []*ParentSubnet{}, parentSubnets)
	})

	t.Run("invalid request", func(t *testing.T) {
		parentSubnets, err := ParentSubnetsFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, parentSubnets)
	})

	t.Run("request", func(t *testing.T) {
		parentSubnet, err := ParentSubnetsFromReader(bytes.NewReader([]byte(
			`[{"ID":"id1", "CIDR":"10.0.0.0/8"}, {"ID":"id2", "CIDR":"192.168.0.0/16"}]`,
		)))
		require.NoError(t, err)
		require.Equal(t, []*ParentSubnet{
			{ID: "id1", CIDR: "10.0.0.0/8"},
			{ID: "id2", CIDR: "192.168.0.0/16"},
		}, parentSubnet)
	})
}
