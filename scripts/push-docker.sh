#!/bin/bash

# Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
# See LICENSE.txt for license information.

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${CIRCLE_TAG:-}" ]]; then
  echo "Pushing lastest for $CIRCLE_BRANCH..."
  TAG=latest
else
  echo "Pushing release $CIRCLE_TAG..."
  TAG="$CIRCLE_TAG"
fi

echo $DOCKER_PASSWORD | docker login --username $DOCKER_USERNAME --password-stdin

docker tag mattermost/genesis:test mattermost/genesis:$TAG

docker push mattermost/genesis:$TAG
