#!/usr/bin/env bash

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
