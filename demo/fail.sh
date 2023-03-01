#!/bin/bash

# create the test namespaces
kubectl apply -f ./demo/namespace.yaml

# lula execute to check for PeerAuthentication resources in our test namespaces.
# validation will fail if they are not present, which in this case they aren't, so we expect validation to fail.
# the output should show 4 resources failing.
./lula execute ./demo/oscal-component.yaml