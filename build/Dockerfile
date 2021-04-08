# Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
# See LICENSE.txt for license information.

# Build the mattermost genesis
ARG DOCKER_BUILD_IMAGE=golang:1.16.3
ARG DOCKER_BASE_IMAGE=alpine:3.13.4

FROM ${DOCKER_BUILD_IMAGE} AS build
WORKDIR /genesis/
COPY . /genesis/
RUN apt-get update -yq && apt-get install -yq unzip
RUN make get-terraform
RUN make build

# Final Image
FROM ${DOCKER_BASE_IMAGE}

LABEL name="Genesis" \
  maintainer="cloud-team@mattermost.com" \
  vendor="Mattermost" \
  distribution-scope="public" \
  architecture="x86_64" \
  url="https://mattermost.com" \
  io.k8s.description="Genesis creates, configures and helps manage AWS accounts, AWS VPCs and VPC Peering connections." \
  io.k8s.display-name="Mattermost Genesis"

ENV GENESIS=/genesis/genesis \
    USER_UID=10001 \
    USER_NAME=genesis

RUN  apk update && apk add libc6-compat && apk add ca-certificates
COPY --from=build /genesis/build/terraform /usr/local/bin/
COPY --from=build /genesis/build/_output/bin/genesis /genesis/genesis
COPY --from=build /genesis/build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup
WORKDIR /genesis/

USER ${USER_UID}

EXPOSE 8075

ENTRYPOINT ["/usr/local/bin/entrypoint"]
