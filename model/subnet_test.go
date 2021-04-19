// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubnetClone(t *testing.T) {
	subnet := &Subnet{
		CIDR: "10.0.0.0/8",
	}

	clone, err := subnet.Clone()
	require.NoError(t, err)
	require.Equal(t, subnet, clone)
}

func TestSubnetFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		subnet, err := SubnetFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, &Subnet{}, subnet)
	})

	t.Run("invalid request", func(t *testing.T) {
		subnet, err := SubnetFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, subnet)
	})

	t.Run("request", func(t *testing.T) {
		subnet, err := SubnetFromReader(bytes.NewReader([]byte(
			`{"ID":"id"}`,
		)))
		require.NoError(t, err)
		require.Equal(t, &Subnet{ID: "id"}, subnet)
	})
}

func TestSubnetsFromReader(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		subnets, err := SubnetsFromReader(bytes.NewReader([]byte(
			``,
		)))
		require.NoError(t, err)
		require.Equal(t, []*Subnet{}, subnets)
	})

	t.Run("invalid request", func(t *testing.T) {
		subnets, err := SubnetsFromReader(bytes.NewReader([]byte(
			`{test`,
		)))
		require.Error(t, err)
		require.Nil(t, subnets)
	})

	t.Run("request", func(t *testing.T) {
		subnet, err := SubnetsFromReader(bytes.NewReader([]byte(
			`[{"ID":"id1", "CIDR":"10.0.0.0/8"}, {"ID":"id2", "CIDR":"192.168.0.0/16"}]`,
		)))
		require.NoError(t, err)
		require.Equal(t, []*Subnet{
			{ID: "id1", CIDR: "10.0.0.0/8"},
			{ID: "id2", CIDR: "192.168.0.0/16"},
		}, subnet)
	})
}
