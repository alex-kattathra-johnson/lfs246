#!/bin/sh

for service in *-service; do
    kubectl apply -f $service.yaml
done