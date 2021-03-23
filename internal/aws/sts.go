package aws

import "github.com/aws/aws-sdk-go/service/sts"

// GetAccountID gets the current AWS Account ID
func (a *Client) GetAccountID() (string, error) {
	callerIdentityOutput, err := a.Service().sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return *callerIdentityOutput.Account, nil
}
