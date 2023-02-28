# Notes for Kyverno querying capabilities

## Prerequisites

- A running kubernetes cluster
  - using k3d for these tests

  ```bash
  k3d cluster create
  ```
- A compiled `lula` binary in the root of the repository

## Configuration and Testing

`oscal-component.yaml` in this directory has been modified to not target pods in a specific namespace, but rather all pods in the cluster.

`namespace.yaml` in this directory has been extended to define multiple namespaces:

- foo
- test
- test1
- test2

`pod.fail.yaml` and `pod.pass.yaml` in this directory have each been extended to define a pod in each test namespace.

***note***: Kyverno will validate any resources that use Pod objects, such as Deployments, Daemonsets, ReplicaSets, Jobs, etc. This will show up in the output as Kyverno validates all of these resources in every namespace.

To see the pods in each namespace fail validation, execute the `fail.sh` script:

***note:*** the output will include pods in already existing namespaces, for example, pods in the `kube-system` namespace will fail because they don't have the label that Kyverno is checking for.

The output should show 0 Passing resources

```bash
./demo/fail.sh
```

To see the pods in each namespace pass validation, execute the `pass.sh` script:

***note:*** this script only adds the proper label to pods in the test namespaces we created. Pods in other namespaces without the proper label will still fail validation.

The output should show 4 Passing resources

```bash
./demo/pass.sh
```
