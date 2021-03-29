package genesis

import (
	"fmt"

	sdkAWS "github.com/aws/aws-sdk-go/aws"
	awstools "github.com/mattermost/genesis/internal/aws"
	model "github.com/mattermost/genesis/model"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// createAccount is used to create new AWS accounts
func createAccount(provisioner *GenProvisioner, account *model.Account, logger *logrus.Entry, awsClient awstools.AWS) error {
	logger.Infof("Creating account %s", account.ID)

	awsCreds, err := awsClient.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", provisioner.controlTowerAccountID, provisioner.controlTowerRole))
	if err != nil {
		return errors.Wrap(err, "failed to assume control tower iam role")
	}

	awsConfig := &sdkAWS.Config{
		Region:      sdkAWS.String(awstools.DefaultAWSRegion),
		Credentials: awsCreds,
		MaxRetries:  sdkAWS.Int(awstools.DefaultAWSClientRetries),
	}
	awsClientControlTower := awstools.NewAWSClientWithConfig(awsConfig, logger)

	if err = awsClientControlTower.ProvisionServiceCatalogProduct(provisioner.ssoUserEmail, provisioner.ssoFirstName, provisioner.ssoLastName, provisioner.managedOU, account); err != nil {
		return errors.Wrap(err, "failed to provision service catalog product")
	}

	wait := 3600
	logger.Infof("Waiting up to %d seconds for account to become ready...", wait)
	if err = awsClientControlTower.WaitForAccountReadiness(account, wait); err != nil {
		logger.WithError(err).Error("failed while waiting for new account to get ready")
		logger.Info("Trying to get AWS Account details")
		if err = awsClientControlTower.GetAccountDetails(account); err != nil {
			return errors.Wrap(err, "failed to get account details")
		}
		return err
	}

	logger.Info("Getting AWS Account physical ID")
	if err = awsClientControlTower.GetAccountDetails(account); err != nil {
		return errors.Wrap(err, "failed to get AWS account details")
	}
	logger.Infof("Creating provisioning IAM role in account %s", account.ProviderMetadataAWS.AWSAccountID)

	logger.Infof("Assuming AWSControlTowerExecution role in destination account %s", account.ProviderMetadataAWS.AWSAccountID)

	genesisAccount, err := awsClient.GetAccountID()
	if err != nil {
		return errors.Wrap(err, "failed to get AWS account physical ID")
	}
	logger.Infof("Code running in account %s", genesisAccount)

	//temp client is used for one time creation of IAM role in destination account that will be used for all future actions in the account
	awsTempCreds, err := awsClientControlTower.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/AWSControlTowerExecution", account.ProviderMetadataAWS.AWSAccountID))
	if err != nil {
		return err
	}
	tempAWSConfig := &sdkAWS.Config{
		Region:      sdkAWS.String(awstools.DefaultAWSRegion),
		Credentials: awsTempCreds,
		MaxRetries:  sdkAWS.Int(awstools.DefaultAWSClientRetries),
	}
	tempDestinationAWSClient := awstools.NewAWSClientWithConfig(tempAWSConfig, logger)

	logger.Infof("Provisioning IAM role in account %s", account.ProviderMetadataAWS.AWSAccountID)
	if err = tempDestinationAWSClient.CreateProvisioningIAMRole(genesisAccount); err != nil {
		return errors.Wrap(err, "failed to create provisioning IAM role in new account")
	}

	return nil
}

// provisionAccount is used to provision AWS accounts
func provisionAccount(provisioner *GenProvisioner, account *model.Account, logger *logrus.Entry, awsClient awstools.AWS) error {
	logger.Infof("Provisioning account %s", account.ID)

	return nil
}

// deleteAccount is used to delete AWS accounts
func deleteAccount(provisioner *GenProvisioner, account *model.Account, logger *logrus.Entry, awsClient awstools.AWS) error {

	logger.Infof("Deleting account with physical id %s", account.ProviderMetadataAWS.AWSAccountID)
	awsCreds, err := awsClient.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", provisioner.controlTowerAccountID, provisioner.controlTowerRole))
	if err != nil {
		return err
	}

	awsConfig := &sdkAWS.Config{
		Region:      sdkAWS.String(awstools.DefaultAWSRegion),
		Credentials: awsCreds,
		MaxRetries:  sdkAWS.Int(awstools.DefaultAWSClientRetries),
	}
	awsClientControlTower := awstools.NewAWSClientWithConfig(awsConfig, logger)

	if err = awsClientControlTower.DeleteServiceCatalogProduct(account.ProviderMetadataAWS.AccountProductID); err != nil {
		return errors.Wrap(err, "failed to delete account")
	}
	return nil
}
