# Lula - The Kubernetes Compliance Engine

[![Go version](https://img.shields.io/github/go-mod/go-version/defenseunicorns/lula?filename=go.mod)](https://go.dev/)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/defenseunicorns/lula/badge)](https://api.securityscorecards.dev/projects/github.com/defenseunicorns/lula)

Lula is a tool written to bridge the gap between expected configuration required for compliance and **_actual_** configuration.

Cloud Native Infrastructure, Platforms, and applications can establish OSCAL documents that live beside source-of-truth code bases. Providing an inheritance model for when a control that the technology can satisfy _IS_ satisfied in a live-environment.

This can be well established and regulated standards such as NIST 800-53. It can also be best practices, Enterprise Standards, or simply team development standards that need to be continuously monitored and validated.

## How does it work?

The primary functionality is leveraging [Kyverno CLI/Engine](https://kyverno.io/docs/kyverno-cli/).
lula:

- Ingests a `oscal-component.yaml` and creates an object in memory
- Queries all `implemented-requirements` for a `rules` field
  - This rules block is a strict port from the rules of a [Kyverno ClusterPolicy](https://kyverno.io/docs/kyverno-policies/) resource
- If a rules field exists:
  - Generate a `ClusterPolicy` resource on the filesystem
  - Execute the `applyCommandHelper` function from Kyverno CLI
    - This will return the number of passing/failing resources in the cluster (or optionally static manifests on the filesystem)
    - If any fail, given valid exclusions that may be present, the control is declared as `Fail`
  - Remove `ClusterPolicy` from the filesystem
  - This is done for each `implemented-requirement` that has a `rules` field
- Generate a report of the findings (`Pass` or `fail` for each control) on the filesystem (optional - can be run with `--dry-run` in order to not write to filesystem)

## Getting Started

## Demo

### Static Manifest Demo

![Resource Demo](./images/resource-demo.gif)


### Live Cluster Demo

![Cluster Demo](./images/cluster-demo.gif)

### Try it out

#### Dependencies

- A running Kubernetes cluster
- GoLang version 1.19.1

#### Steps

1. Clone the repository to your local machine and change into the `lula` directory

    ```bash
    git clone https://github.com/defenseunicorns/lula.git && cd lula
    ```

1. While in the `lula` directory, compile the tool into an executable binary. This outputs the `lula` binary to the current working directory.

    ```bash
    go build .
    ```

1. Apply the `./demo/namespace.yaml` file to create a namespace for the demo

    ```bash
    kubectl apply -f ./demo/namespace.yaml
    ```

1. Apply the `./demo/pod.fail.yaml` to create a pod in your cluster

    ```bash
    kubectl apply -f ./demo/pod.fail.yaml
    ```

1. Run the following command in the `lula` directory:

    ```bash
    ./lula validate ./demo/oscal-component.yaml
    ```

    The output in your terminal should inform you that there is at least one failing pod in the cluster:

    ```bash
    Applying 1 policy rule to 19 resources...

    policy 42c2ffdc-5f05-44df-a67f-eec8660aeffd -> resource foo/Pod/demo-pod failed: 
    1. ID-1: validation error: Every pod in namespace 'foo' should have 'foo=bar' label. rule ID-1 failed at path /metadata/labels/foo/ 
    UUID: 42C2FFDC-5F05-44DF-A67F-EEC8660AEFFD
            Resources Passing: 0
            Resources Failing: 1
            Status: Fail
    ```

1. Now, apply the `./demo/pod.pass.yaml` file to your cluster to configure the pod to pass compliance validation:

    ```bash
    kubectl apply -f ./demo/pod.pass.yaml
    ```

1. Run the following command in the `lula` directory:

    ```bash
    ./lula validate ./demo/oscal-component.yaml
    ```

    The output should now show the pod as passing the compliance requirement:

    ```bash
    Applying 1 policy rule to 19 resources...
    UUID: 42C2FFDC-5F05-44DF-A67F-EEC8660AEFFD
            Resources Passing: 1
            Resources Failing: 0
            Status: Pass
    ```

## Future Extensibility

- Support for cloud infrastructure state queries
- Support for API validation

## Developing

- Go 1.19
