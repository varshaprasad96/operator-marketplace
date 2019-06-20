#!/bin/bash
# This script downloads and runs the OpenShift installer associated with the provided tag.

echo 'Delete exsisting marketplace'
oc delete deployment marketplace-operator

echo 'Delete exsisting prometheus'
oc delete prometheus prometheus

echo 'Redeploying marketplace'
oc apply -f manifests/10_operator.yaml

echo 'Deploy service-object'
oc apply -f service-object.yaml

echo 'Deploy service-monitor'
oc apply -f service-monitor.yaml

echo 'Deploy prometheus'
oc apply -f prometheus.yaml
