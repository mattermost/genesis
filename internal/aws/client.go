// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package aws

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/aws/aws-sdk-go/service/ram/ramiface"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/aws/aws-sdk-go/service/servicecatalog/servicecatalogiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	log "github.com/sirupsen/logrus"
)

// AWS interface for use by other packages.
type AWS interface {
	GetAccountAliases() (*iam.ListAccountAliasesOutput, error)
	GetCloudEnvironmentName() (string, error)
	AssumeRole(roleArn string) (*credentials.Credentials, error)
	GetAccountID() (string, error)
	AssociateTGWShare(resourceShareARN, principalID string) error
	DisassociateTGWShare(resourceShareARN, principalID string) error
}

// NewAWSClientWithConfig returns a new instance of Client with a custom configuration.
func NewAWSClientWithConfig(config *aws.Config, logger log.FieldLogger) *Client {
	return &Client{
		logger: logger,
		config: config,
		mux:    &sync.Mutex{},
	}
}

// Service hold AWS clients for each service.
type Service struct {
	ec2            ec2iface.EC2API
	iam            iamiface.IAMAPI
	sts            stsiface.STSAPI
	serviceCatalog servicecatalogiface.ServiceCatalogAPI
	ram            ramiface.RAMAPI
}

// NewService creates a new instance of Service.
func NewService(sess *session.Session) *Service {
	return &Service{
		iam:            iam.New(sess),
		ec2:            ec2.New(sess),
		sts:            sts.New(sess),
		serviceCatalog: servicecatalog.New(sess),
		ram:            ram.New(sess),
	}
}

// Client is a client for interacting with AWS resources.
type Client struct {
	logger  log.FieldLogger
	service *Service
	config  *aws.Config
	mux     *sync.Mutex
}

// Service contructs an AWS session if not yet successfully done and returns AWS clients.
func (c *Client) Service() *Service {
	if c.service == nil {
		sess, err := NewAWSSessionWithLogger(c.config, c.logger.WithField("tools-aws", "client"))
		if err != nil {
			c.logger.WithError(err).Error("failed to initialize AWS session")
			// Calls to AWS will fail until a healthy session is acquired.
			return NewService(&session.Session{})
		}

		c.mux.Lock()
		c.service = NewService(sess)
		c.mux.Unlock()
	}

	return c.service
}
