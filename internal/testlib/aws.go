// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package testlib

import (
	"github.com/golang/mock/gomock"
	mocks "github.com/mattermost/genesis/internal/mocks/aws-sdk"
)

// AWSMockedAPI has all AWS mocked services. New services should be added here.
type AWSMockedAPI struct {
	IAM            *mocks.MockIAMAPI
	EC2            *mocks.MockEC2API
	STS            *mocks.MockSTSAPI
	ServiceCatalog *mocks.MockServiceCatalogAPI
	RAM            *mocks.MockRAMAPI
}

// NewAWSMockedAPI returns an instance of AWSMockedAPI.
func NewAWSMockedAPI(ctrl *gomock.Controller) *AWSMockedAPI {
	return &AWSMockedAPI{
		IAM:            mocks.NewMockIAMAPI(ctrl),
		EC2:            mocks.NewMockEC2API(ctrl),
		STS:            mocks.NewMockSTSAPI(ctrl),
		ServiceCatalog: mocks.NewMockServiceCatalogAPI(ctrl),
		RAM:            mocks.NewMockRAMAPI(ctrl),
	}
}
