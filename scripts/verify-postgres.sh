#!/usr/bin/env bash

# Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
# See LICENSE.txt for license information.

set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

for i in `seq 1 10`;
do
    nc -z localhost 5432 && echo Success && exit 0
    echo -n .
    sleep 1
done
echo Failed waiting for Postgres && exit 1
