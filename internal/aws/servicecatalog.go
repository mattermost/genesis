package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
)

// WaitForAccountReadiness is checking if a new account is in ready state.
func (a *Client) WaitForAccountReadiness(account *model.Account, timeout int) error {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return errors.New("timed out waiting for account to become ready")
		default:
			ready, err := a.ValidateAccount(account)
			if err != nil {
				return err
			}
			if ready {
				return nil
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// ValidateAccount checks if an AWS account is in status available state.
func (a *Client) ValidateAccount(account *model.Account) (bool, error) {
	products, err := a.Service().serviceCatalog.ScanProvisionedProducts(&servicecatalog.ScanProvisionedProductsInput{})
	if err != nil {
		return false, err
	}

	for _, product := range products.ProvisionedProducts {
		if *product.Name == account.ID && *product.Status == servicecatalog.StatusAvailable {
			return true, nil
		}
	}

	return false, nil
}

// GetAccountDetails returns the details of an AWS account, such as account physical and product ID.
func (a *Client) GetAccountDetails(account *model.Account) error {
	products, err := a.Service().serviceCatalog.SearchProvisionedProducts(&servicecatalog.SearchProvisionedProductsInput{})
	if err != nil {
		return err
	}
	for _, product := range products.ProvisionedProducts {
		if *product.Name == account.ID {
			account.ProviderMetadataAWS.AWSAccountID = *product.PhysicalId
			account.ProviderMetadataAWS.AccountProductID = *product.Id
		}
	}

	return nil
}

// GetProvisioningArtifactID returns the current active Service Catalog provisioning artifact ID.
func (a *Client) GetProvisioningArtifactID(productID string) (string, error) {
	product, err := a.Service().serviceCatalog.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(productID),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to get the service catalog product details")
	}

	for _, pa := range product.ProvisioningArtifactSummaries {
		provisioningArtifact, err := a.Service().serviceCatalog.DescribeProvisioningArtifact(&servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(*pa.Id),
		})
		if err != nil {
			return "", errors.Wrap(err, "failed to get the service catalog provisioning artifact details")
		}
		if *provisioningArtifact.ProvisioningArtifactDetail.Active {
			return *pa.Id, nil
		}
	}
	return "", errors.Errorf("failed to get the active service catalog provisioning artifact")
}

// ProvisionProduct calls the AWS API to provision a new service catalog product.
func (a *Client) ProvisionProduct(input servicecatalog.ProvisionProductInput) (*servicecatalog.ProvisionProductOutput, error) {
	product, err := a.Service().serviceCatalog.ProvisionProduct(&input)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// GetProvisioningArtifactID returns the current active Service Catalog provisioning artifact ID.
func GetProvisioningArtifactID(servicecatalogService *servicecatalog.ServiceCatalog, productID string) (string, error) {
	product, err := servicecatalogService.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(productID),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to get the service catalog product details")
	}

	for _, pa := range product.ProvisioningArtifactSummaries {
		provisioningArtifact, err := servicecatalogService.DescribeProvisioningArtifact(&servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(*pa.Id),
		})
		if err != nil {
			return "", errors.Wrap(err, "failed to get the service catalog provisioning artifact details")
		}
		if *provisioningArtifact.ProvisioningArtifactDetail.Active {
			return *pa.Id, nil
		}
	}
	return "", errors.Errorf("failed to get the active service catalog provisioning artifact")
}

// ProvisionServiceCatalogProduct handles the steps to provision a new service catalog product.
func (a *Client) ProvisionServiceCatalogProduct(ssoUserEmail, ssoFirstName, ssoLastName, managedOU string, account *model.Account) error {
	provisioningArtifactID, err := a.GetProvisioningArtifactID(account.ProviderMetadataAWS.ServiceCatalogProductID)
	if err != nil {
		return err
	}

	accountInput := servicecatalog.ProvisionProductInput{
		ProductId:              aws.String(account.ProviderMetadataAWS.ServiceCatalogProductID),
		ProvisioningArtifactId: aws.String(provisioningArtifactID),
		ProvisionedProductName: aws.String(account.ID),
		ProvisioningParameters: []*servicecatalog.ProvisioningParameter{
			{
				Key:   aws.String("SSOUserEmail"),
				Value: aws.String(ssoUserEmail),
			},
			{
				Key:   aws.String("SSOUserFirstName"),
				Value: aws.String(ssoFirstName),
			},
			{
				Key:   aws.String("SSOUserLastName"),
				Value: aws.String(ssoLastName),
			},
			{
				Key:   aws.String("ManagedOrganizationalUnit"),
				Value: aws.String(managedOU),
			},
			{
				Key:   aws.String("AccountName"),
				Value: aws.String(fmt.Sprintf("%s-%s", AccountProductPrefix, account.ID)),
			},
			{
				Key:   aws.String("AccountEmail"),
				Value: aws.String(fmt.Sprintf("%s+%s@mattermost.com", AccountEmailPrefix, account.ID[:5])),
			},
		},
	}
	_, err = a.ProvisionProduct(accountInput)
	if err != nil && IsErrorCode(err, servicecatalog.ErrCodeDuplicateResourceException) {
		a.logger.Info("Service catalog product already provisioned, skipping...")
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

// DeleteServiceCatalogProduct deletes a service catalog product.
func (a *Client) DeleteServiceCatalogProduct(productID string) error {
	_, err := a.Service().serviceCatalog.TerminateProvisionedProduct(&servicecatalog.TerminateProvisionedProductInput{
		ProvisionedProductId: aws.String(productID),
	})
	if err != nil {
		return err
	}
	return nil
}
