// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package model_test

import (
	"testing"

	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/assert"
)

func TestAddParentSubnetRequestValid(t *testing.T) {
	var testCases = []struct {
		testName     string
		request      *model.AddParentSubnetRequest
		requireError bool
	}{
		{"defaults", &model.AddParentSubnetRequest{CIDR: "10.0.0.0/8"}, false},
		{"empty CIDR", &model.AddParentSubnetRequest{CIDR: ""}, true},
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
