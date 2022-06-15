#!/bin/bash

kubectl delete -f aggregator.yml
docker build ../../../ -f ../cpu_server/Dockerfile_aggregator -t iot_aggregator
docker tag iot_aggregator localhost:32000/iot_aggregator
docker push localhost:32000/iot_aggregator
kubectl apply -f aggregator.yml
kubectl apply -f influxdb.yml
kubectl apply -f mqtt.yml

# This assumes you already have Prometheus deployed
