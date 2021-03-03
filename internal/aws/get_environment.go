package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/pkg/errors"
)

// GetCloudEnvironmentName looks for a standard cloud account environment name
// and returns it.
func (c *Client) GetCloudEnvironmentName() (string, error) {
	accountAliases, err := c.GetAccountAliases()
	if err != nil {
		return "", errors.Wrap(err, "failed to get account aliases")
	}
	if len(accountAliases.AccountAliases) < 1 {
		return "", errors.New("account alias not defined")
	}

	for _, alias := range accountAliases.AccountAliases {
		if strings.HasPrefix(*alias, "mattermost-cloud") && len(strings.Split(*alias, "-")) == 3 {
			envName := strings.Split(*alias, "-")[2]
			if len(envName) == 0 {
				return "", errors.New("environment name value was empty")
			}

			return envName, nil
		}
	}

	return "", errors.New("account environment name could not be found from account aliases")
}

// GetAccountAliases returns the AWS account name aliases.
func (a *Client) GetAccountAliases() (*iam.ListAccountAliasesOutput, error) {
	accountAliases, err := a.Service().iam.ListAccountAliases(&iam.ListAccountAliasesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get AWS account name aliases")
	}
	return accountAliases, nil
}
