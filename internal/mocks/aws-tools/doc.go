// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

// Package mocks to create the mocks run go generate to regenerate this package.
//go:generate ../../../bin/mockgen -package=mocks -destination ./client.go -source ../../aws/client.go
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt client.go > _client.go && mv _client.go client.go"
package mocks //nolint
