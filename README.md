# Compliance-Auditor

Compliance Auditor is a tool written to bridge the gap between expected configruation required for compliance and **_actual_** configuration.

Cloud Native Infrastructure, Platforms, and applications can establish OSCAL documents that live beside source-of-truth code bases. Providing an inheritance model for when a control that the technology can satisfy _IS_ satisfied in a live-environment. 

This can be well established and regulated standards such as NIST 800-53. It can also be best practices, Enterprise Standards, or simply team development standards that need to be continuously monitored and validated.

## Proof of Concept
With an established cluster, ingest an OSCAL document for Istio which contains a validation field for _WHEN_ a control (AC-4) is satisfied. Execute a query against the established cluster and produce an OSCAL document that states whether the control is passing or failing.

Cluster will be postured to fail the first run. The demonstration application will then be configured for istio-injection and be applied to the cluster. The tool will be executed and the output will validate that the control is now satisfied.

## Current Investigation
- Can we write a wrapper around Kyverno for validating resource state?
    - Example: All pods (excluding kube-system namespace) must have at-minimum container with the image istio-proxy (meaning that they are istio-injected)
        - This _could_ be written as a kyverno policy with ease
    - Ingested query would be a `match/validate` policy-like object(s)
    - For Each query, Execute an `apply` against a live-cluster, processing whether the policy-like object pass/fail
    - Generate OSCAL with results

## Extensibility
- Support for cloud infrastructure state queries
- Support for API validation

## Getting Started

## Developing
- GO 1.19