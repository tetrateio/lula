#!/bin/bash

make build

k3d cluster delete && k3d cluster create
