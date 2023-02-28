#!/bin/bash

# create test namespaces
kubectl apply -f ./demo/namespace.yaml

# create failing test pod in each test namespace
kubectl apply -f ./demo/pod.fail.yaml

# wait for test pods to be ready
for namespace in "foo" "test" "test1" "test2"
do
    kubectl wait --for=condition=Ready pod/demo-pod -n "${namespace}"
done

# lula execute to validate pass/fail status
./lula execute ./demo/oscal-component.yaml
