package genesis_test

import (
	"testing"

	genesis "github.com/mattermost/genesis/internal/genesis"
	model "github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/require"
)

func TestSplitParentSubnet(t *testing.T) {
	t.Run("split cidr plus one range", func(t *testing.T) {
		// logger := testlib.MakeLogger(t)
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 9,
			CreateAt:   10,
		}
		subnets, err := genesis.SplitParentSubnet(parentSubnet1)
		require.NoError(t, err)

		actualSubnets := []model.Subnet{
			{
				CIDR:           "10.0.0.0/9",
				Used:           false,
				ParentSubnet:   "10.0.0.0/8",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.128.0.0/9",
				Used:           false,
				ParentSubnet:   "10.0.0.0/8",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
		}
		require.Equal(t, subnets, actualSubnets)
	})
	t.Run("split cidr plus two range", func(t *testing.T) {
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 10,
			CreateAt:   10,
		}
		subnets, err := genesis.SplitParentSubnet(parentSubnet1)
		require.NoError(t, err)

		actualSubnets := []model.Subnet{
			{
				CIDR:           "10.0.0.0/10",
				Used:           false,
				ParentSubnet:   "10.0.0.0/8",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.64.0.0/10",
				Used:           false,
				ParentSubnet:   "10.0.0.0/8",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.128.0.0/10",
				Used:           false,
				ParentSubnet:   "10.0.0.0/8",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.192.0.0/10",
				Used:           false,
				ParentSubnet:   "10.0.0.0/8",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
		}
		require.Equal(t, subnets, actualSubnets)
	})
	t.Run("split cidr plus three range", func(t *testing.T) {
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/20",
			SplitRange: 23,
			CreateAt:   10,
		}
		subnets, err := genesis.SplitParentSubnet(parentSubnet1)
		require.NoError(t, err)

		actualSubnets := []model.Subnet{
			{
				CIDR:           "10.0.0.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.2.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.4.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.6.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.8.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.10.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.12.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
			{
				CIDR:           "10.0.14.0/23",
				Used:           false,
				ParentSubnet:   "10.0.0.0/20",
				SubnetMetadata: &model.SubnetMetadata{},
				CreateAt:       10,
			},
		}
		require.Equal(t, subnets, actualSubnets)
	})
	t.Run("split cidr invalid range", func(t *testing.T) {
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 33,
			CreateAt:   10,
		}
		_, err := genesis.SplitParentSubnet(parentSubnet1)
		require.Error(t, err)
	})
	t.Run("split cidr invalid cidr", func(t *testing.T) {
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0//8",
			SplitRange: 10,
			CreateAt:   10,
		}
		_, err := genesis.SplitParentSubnet(parentSubnet1)
		require.Error(t, err)
	})
	t.Run("check returner range", func(t *testing.T) {
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 10,
			CreateAt:   10,
		}
		subnets, err := genesis.SplitParentSubnet(parentSubnet1)
		require.NoError(t, err)

		require.Equal(t, 4, len(subnets))
	})
	t.Run("check returned range of higher split", func(t *testing.T) {
		parentSubnet1 := &model.ParentSubnet{
			CIDR:       "10.0.0.0/8",
			SplitRange: 18,
			CreateAt:   10,
		}
		subnets, err := genesis.SplitParentSubnet(parentSubnet1)
		require.NoError(t, err)

		require.Equal(t, 1024, len(subnets))
	})
}
