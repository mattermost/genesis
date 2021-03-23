package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

// AssumeRoleCredentialsProvider describes assume role credentials.
type AssumeRoleCredentialsProvider struct {
	AssumeRoleCredentials *sts.Credentials
}

// CreateProvisioningIAMRole is used to crate the provisioning role in new accounts.
func (a *Client) CreateProvisioningIAMRole(trustAccountID string) error {
	_, err := a.Service().iam.CreateRole(&iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(fmt.Sprintf("{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Principal\": {\n        \"AWS\": \"arn:aws:iam::%s:root\"\n      },\n      \"Action\": \"sts:AssumeRole\"\n    }\n  ]\n}", trustAccountID)),
		Description:              aws.String("This is the provisioning Role. Will be used by Genesis and other applications to provision the account."),
		RoleName:                 aws.String(AccountProvisioningRoleName),
	})
	if err != nil {
		return err
	}
	return nil
}

// AssumeRole assumes an IAM role using local credentials and returns credentials.
func (a *Client) AssumeRole(roleArn string) (*credentials.Credentials, error) {
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("AssumeRoleSession"),
	}

	assumeRole, err := a.Service().sts.AssumeRole(input)
	if err != nil {
		return nil, err
	}
	provider := NewAssumeRoleCredentialsProvider(assumeRole.Credentials)

	return credentials.NewCredentials(provider), nil
}

// NewAssumeRoleCredentialsProvider returns AssumeRoleCredentialsProvider using provided credentials.
func NewAssumeRoleCredentialsProvider(credentials *sts.Credentials) *AssumeRoleCredentialsProvider {
	return &AssumeRoleCredentialsProvider{
		AssumeRoleCredentials: credentials,
	}
}

// Retrieve returns the creds values.
func (c AssumeRoleCredentialsProvider) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     *c.AssumeRoleCredentials.AccessKeyId,
		SecretAccessKey: *c.AssumeRoleCredentials.SecretAccessKey,
		SessionToken:    *c.AssumeRoleCredentials.SessionToken,
		ProviderName:    "AssumeRoleCredentialsProvider",
	}, nil

}

// IsExpired checks if the assume role session has expired.
func (c AssumeRoleCredentialsProvider) IsExpired() bool {
	return c.AssumeRoleCredentials.Expiration.After(time.Now())
}
