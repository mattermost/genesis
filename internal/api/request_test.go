// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package api

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/mattermost/genesis/model"
	"github.com/stretchr/testify/require"
)

func TestNewCreateAccountRequestFromReader(t *testing.T) {
	defaultCreateAccountRequest := func() *model.CreateAccountRequest {
		return &model.CreateAccountRequest{
			Provider:                "aws",
			ServiceCatalogProductID: "prod-12345",
		}
	}

	t.Run("invalid request", func(t *testing.T) {
		accountRequest, err := model.NewCreateAccountRequestFromReader(bytes.NewReader([]byte(
			`{`,
		)))
		require.Error(t, err)
		require.Nil(t, accountRequest)
	})

	t.Run("unsupported provider", func(t *testing.T) {
		accountRequest, err := model.NewCreateAccountRequestFromReader(bytes.NewReader([]byte(
			`{"Provider": "azure"}`,
		)))
		require.EqualError(t, err, "create account request failed validation: unsupported provider azure")
		require.Nil(t, accountRequest)
	})

	t.Run("partial request", func(t *testing.T) {
		accountRequest, err := model.NewCreateAccountRequestFromReader(bytes.NewReader([]byte(
			`{"ServiceCatalogProductID": "prod-12345"}`,
		)))
		require.NoError(t, err)
		modifiedDefaultCreateAccountRequest := defaultCreateAccountRequest()
		modifiedDefaultCreateAccountRequest.ServiceCatalogProductID = "prod-12345"
		require.Equal(t, modifiedDefaultCreateAccountRequest, accountRequest)
	})

	t.Run("full request", func(t *testing.T) {
		accountRequest, err := model.NewCreateAccountRequestFromReader(bytes.NewReader([]byte(
			`{"Provider": "aws", "ServiceCatalogProductID": "prod-12345", "Provision": true}`,
		)))
		require.NoError(t, err)
		require.Equal(t, &model.CreateAccountRequest{
			Provider:                model.ProviderAWS,
			ServiceCatalogProductID: "prod-12345",
			Provision:               true,
		}, accountRequest)
	})
}

func TestGetAccountsRequestApplyToURL(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		u, err := url.Parse("http://localhost:8073")
		require.NoError(t, err)

		getAccountsRequest := &model.GetAccountsRequest{}
		getAccountsRequest.ApplyToURL(u)

		require.Equal(t, "page=0&per_page=0", u.RawQuery)
	})

	t.Run("changes", func(t *testing.T) {
		u, err := url.Parse("http://localhost:8073")
		require.NoError(t, err)

		getAccountsRequest := &model.GetAccountsRequest{
			Page:           10,
			PerPage:        123,
			IncludeDeleted: true,
		}
		getAccountsRequest.ApplyToURL(u)

		require.Equal(t, "include_deleted=true&page=10&per_page=123", u.RawQuery)
	})
}
