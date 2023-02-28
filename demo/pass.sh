#!/bin/bash

# create test namespaces
kubectl apply -f ./demo/namespace.yaml

# add correct label to test pod in each test namespace to get validation to pass
# pods in other namespaces that don't have the label will still validation
kubectl apply -f ./demo/pod.pass.yaml

# wait for test pods to be ready
for namespace in "foo" "test" "test1" "test2"
do
    kubectl wait --for=condition=Ready pod/demo-pod -n "${namespace}"
done

# lula execute to validate pass/fail status
./lula execute ./demo/oscal-component.yaml
