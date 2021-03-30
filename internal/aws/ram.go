package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/pkg/errors"
)

func (a *Client) AssociateTGWShare(resourceShareARN, principalID string) error {
	if _, err := a.Service().ram.AssociateResourceShare(&ram.AssociateResourceShareInput{
		Principals:       []*string{aws.String(principalID)},
		ResourceShareArn: aws.String(resourceShareARN),
	}); err != nil {
		return errors.Wrap(err, "failed to associate transit gateway with account")
	}

	return nil
}

func (a *Client) DisassociateTGWShare(resourceShareARN, principalID string) error {
	if _, err := a.Service().ram.DisassociateResourceShare(&ram.DisassociateResourceShareInput{
		Principals:       []*string{aws.String(principalID)},
		ResourceShareArn: aws.String(resourceShareARN),
	}); err != nil {
		return errors.Wrap(err, "failed to disassociate transit gateway with account")
	}

	return nil
}
