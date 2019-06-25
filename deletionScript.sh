#!/bin/bash
# The script helps in deleting deployment, svc, routes with name "marketplace-operator"

echo 'Deleting deployment'
oc delete deployment marketplace-operator

echo 'Deleteing svc'
oc delete svc marketplace-operator

echo 'Deleting routes'
oc delete routes marketplace-operator


echo 'Showing the outputs'
echo 'current deployments'
oc get deployments

echo 'current services'
oc get svc

echo 'current routes'
oc get routes
