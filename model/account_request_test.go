// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model_test

import (
	"testing"

	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccountRequestValid(t *testing.T) {
	var testCases = []struct {
		testName     string
		request      *model.CreateAccountRequest
		requireError bool
	}{
		{"defaults", &model.CreateAccountRequest{ServiceCatalogProductID: "prod-12345"}, false},
		{"invalid provider", &model.CreateAccountRequest{Provider: "blah"}, true},
		{"invalid service catalog product id", &model.CreateAccountRequest{ServiceCatalogProductID: ""}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			tc.request.SetDefaults()

			if tc.requireError {
				assert.Error(t, tc.request.Validate())
			} else {
				assert.NoError(t, tc.request.Validate())
			}
		})
	}
}
