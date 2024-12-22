#!/bin/sh

set -ex

export VERSION=v0.0.1

for service in *-service; do
    docker build --build-arg SERVICE=$service -t localhost:5001/$service:$VERSION .
    docker push localhost:5001/$service:$VERSION
done