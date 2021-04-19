// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

// Package mocks to create the mocks run go generate to regenerate this package.
//go:generate ../../../bin/mockgen -package=mocks -destination ./ec2.go github.com/aws/aws-sdk-go/service/ec2/ec2iface EC2API
//go:generate ../../../bin/mockgen -package=mocks -destination ./iam.go github.com/aws/aws-sdk-go/service/iam/iamiface IAMAPI
//go:generate ../../../bin/mockgen -package=mocks -destination ./sts.go github.com/aws/aws-sdk-go/service/sts/stsiface STSAPI
//go:generate ../../../bin/mockgen -package=mocks -destination ./service_catalog.go github.com/aws/aws-sdk-go/service/servicecatalog/servicecatalogiface ServiceCatalogAPI
//go:generate ../../../bin/mockgen -package=mocks -destination ./ram.go github.com/aws/aws-sdk-go/service/ram/ramiface RAMAPI

//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt ec2.go > _ec2.go && mv _ec2.go ec2.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt iam.go > _iam.go && mv _iam.go iam.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt sts.go > _sts.go && mv _sts.go sts.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt service_catalog.go > _service_catalog.go && mv _service_catalog.go service_catalog.go"
//go:generate /usr/bin/env bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt ram.go > _ram.go && mv _ram.go ram.go"
package mocks //nolint
