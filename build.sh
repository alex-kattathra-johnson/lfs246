#!/bin/sh

set -ex

export VERSION=v0.1.2

for service in *-service; do
    docker build --build-arg SERVICE=$service -t localhost:5001/$service:$VERSION .
    docker push localhost:5001/$service:$VERSION
done