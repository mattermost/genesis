// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model

// SubnetMetadata is the subnet metadata.
type SubnetMetadata struct {
	AccountID  string
	VPCID      string
	VPCPeering bool
}
