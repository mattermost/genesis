#!/bin/bash

# Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
# See LICENSE.txt for license information.

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

export TAG="${CIRCLE_SHA1:0:7}"

echo $DOCKER_PASSWORD | docker login --username $DOCKER_USERNAME --password-stdin

docker tag mattermost/genesis:test mattermost/genesis:$TAG

docker push mattermost/genesis:$TAG
