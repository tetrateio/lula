#!/bin/bash

kubectl apply -f ./demo/namespace.yaml

kubectl apply -f ./demo/pod.fail.yaml
