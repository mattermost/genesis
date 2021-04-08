package genesis

import (
	"fmt"

	sdkAWS "github.com/aws/aws-sdk-go/aws"
	awstools "github.com/mattermost/genesis/internal/aws"
	terraform "github.com/mattermost/genesis/internal/terraform"
	model "github.com/mattermost/genesis/model"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

// createAccount is used to create new AWS accounts
func createAccount(provisioner *GenProvisioner, account *model.Account, logger *logrus.Entry, awsClient awstools.AWS) error {
	logger.Infof("Creating account %s", account.ID)

	awsCreds, err := awsClient.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", provisioner.accountCreation.ControlTowerAccountID, provisioner.accountCreation.ControlTowerRole))
	if err != nil {
		return errors.Wrap(err, "failed to assume control tower iam role")
	}

	awsConfig := &sdkAWS.Config{
		Region:      sdkAWS.String(awstools.DefaultAWSRegion),
		Credentials: awsCreds,
		MaxRetries:  sdkAWS.Int(awstools.DefaultAWSClientRetries),
	}
	awsClientControlTower := awstools.NewAWSClientWithConfig(awsConfig, logger)

	if err = awsClientControlTower.ProvisionServiceCatalogProduct(provisioner.accountCreation.SSOUserEmail, provisioner.accountCreation.SSOFirstName, provisioner.accountCreation.SSOLastName, provisioner.accountCreation.ManagedOU, account); err != nil {
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

	logger.Infof("Attaching IAM policy in account %s", account.ProviderMetadataAWS.AWSAccountID)
	if err = tempDestinationAWSClient.AttachIAMPolicy(genesisAccount); err != nil {
		return errors.Wrap(err, "failed to attach IAM policy to provisioning IAM role")
	}

	return nil
}

// provisionAccount is used to provision AWS accounts
func provisionAccount(provisioner *GenProvisioner, account *model.Account, logger *logrus.Entry, awsClient awstools.AWS) error {
	logger.Infof("Provisioning account %s", account.ID)

	logger.Infof("Associating account %s with TGW share", account.ProviderMetadataAWS.AWSAccountID)
	awsCreds, err := awsClient.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", provisioner.accountProvision.CoreAccountID, awstools.TGWShareAssociationRole))
	if err != nil {
		return errors.Wrap(err, "failed to assume core account iam role")
	}

	awsConfig := &sdkAWS.Config{
		Region:      sdkAWS.String(awstools.DefaultAWSRegion),
		Credentials: awsCreds,
		MaxRetries:  sdkAWS.Int(awstools.DefaultAWSClientRetries),
	}
	CoreAWSClient := awstools.NewAWSClientWithConfig(awsConfig, logger)
	resourceShareARN := fmt.Sprintf("arn:aws:ram:us-east-1:%s:resource-share/%s", provisioner.accountProvision.CoreAccountID, provisioner.accountProvision.ResourceShareID)

	if err = CoreAWSClient.AssociateTGWShare(resourceShareARN, account.ProviderMetadataAWS.AWSAccountID); err != nil {
		return errors.Wrap(err, "failed to associate TGW share with the AWS account")
	}

	tf, err := terraform.New("terraform/aws/networking", provisioner.accountProvision.StateBucket, logger)
	if err != nil {
		return errors.Wrap(err, "failed to initiate Terraform")
	}

	err = tf.Init(account.ID)
	if err != nil {
		return errors.Wrap(err, "failed to run Terraform init")
	}

	logger.Infof("Applying Terraform template with VPC %s deployment", account.AccountMetadata.Subnet)
	if err = tf.Apply(provisioner.accountProvision, account.AccountMetadata.Subnet, account.ProviderMetadataAWS.AWSAccountID); err != nil {
		return errors.Wrap(err, "failed to run Terraform apply")
	}
	logger.Info("Successfully ran Terraform apply")
	return nil
}

// deleteAccount is used to delete AWS accounts
func deleteAccount(provisioner *GenProvisioner, account *model.Account, logger *logrus.Entry, awsClient awstools.AWS) error {

	tf, err := terraform.New("terraform/aws/networking", provisioner.accountProvision.StateBucket, logger)
	if err != nil {
		return errors.Wrap(err, "failed to initiate Terraform")
	}

	err = tf.Init(account.ID)
	if err != nil {
		return errors.Wrap(err, "failed to run Terraform init")
	}

	logger.Info("Destroying Terraform resources")
	err = tf.Destroy(account.ProviderMetadataAWS.AWSAccountID)
	if err != nil {
		return errors.Wrap(err, "failed to run Terraform destroy")
	}
	logger.Info("Successfully destroyed Terraform resources")

	logger.Infof("Deleting account with physical id %s", account.ProviderMetadataAWS.AWSAccountID)
	awsCreds, err := awsClient.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", provisioner.accountCreation.ControlTowerAccountID, provisioner.accountCreation.ControlTowerRole))
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

	if account.AccountMetadata.Provision {
		logger.Infof("Disassociating account %s with TGW share", account.ProviderMetadataAWS.AWSAccountID)

		coreAWSCreds, err := awsClient.AssumeRole(fmt.Sprintf("arn:aws:iam::%s:role/%s", provisioner.accountProvision.CoreAccountID, awstools.TGWShareAssociationRole))
		if err != nil {
			return errors.Wrap(err, "failed to assume core account iam role")
		}

		coreAWSConfig := &sdkAWS.Config{
			Region:      sdkAWS.String(awstools.DefaultAWSRegion),
			Credentials: coreAWSCreds,
			MaxRetries:  sdkAWS.Int(awstools.DefaultAWSClientRetries),
		}
		CoreAWSClient := awstools.NewAWSClientWithConfig(coreAWSConfig, logger)
		resourceShareARN := fmt.Sprintf("arn:aws:ram:us-east-1:%s:resource-share/%s", provisioner.accountProvision.CoreAccountID, provisioner.accountProvision.ResourceShareID)

		if err = CoreAWSClient.DisassociateTGWShare(resourceShareARN, account.ProviderMetadataAWS.AWSAccountID); err != nil {
			return errors.Wrap(err, "failed to disassociate TGW share with the AWS account")
		}
	}

	return nil
}
