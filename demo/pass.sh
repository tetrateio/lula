#!/bin/bash

# create a PeerAuthentication resource in each test namespace
kubectl apply -f ./demo/peer-auth.yaml

# lula execute to check for the PeerAuthentication resources.
# the output should show 4 resources passing.
./lula execute ./demo/oscal-component.yaml
