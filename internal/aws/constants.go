// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package aws

const (
	// DefaultAWSRegion is the default AWS region for AWS resources.
	DefaultAWSRegion = "us-east-1"

	// AccountProvisioningRoleName is the name of the provisioning role that
	// will be used by Genesis and other applications to provision the account.
	AccountProvisioningRoleName = "MattermostAccountProvisioningRole"

	// AccountProductPrefix is the prefix of all account products
	AccountProductPrefix = "cloud-enterprise"

	// AccountEmailPrefix is the prefix of the email created for each account product
	AccountEmailPrefix = "cloud-team"

	// DefaultAWSClientRetries supplies how many time the AWS client will
	// retry a failed call.
	DefaultAWSClientRetries = 3

	// The name of the IAM role to use for TGW share associations.
	TGWShareAssociationRole = "tgw-share-association-role"
)
