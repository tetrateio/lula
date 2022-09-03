# Questions and Answers
This document captures conversations that provide insight into capabiltiies, roles, responsibilities, and functionality of the tool.
These are all subject to change and simply an attempt to document nuances across initial creation.

### What is the intent for the tool?
Command Line Interface tool for ingesting OSCAL documents that declare _WHEN_ a control is satisfied and executing queries against a given technology for validation of said control.

### CLI or Event-Driven
CLI is initial focus (human in the loop) - with either a cronjob orchestration wrapper OR event-driven controller deployment.

### How will the tool query kubernetes clusters?
(investigating) Kyverno provides an engine that could provide a well documented input (queries as policy-like objects) and mature logic for testing if a given policy would pass or fail

### How will the tool aggregate many documents, across mutliple technologies?
The primary goal of the tool is to develop with with extensibility in mind. Aggregation through technology flags for targeted input OSCAL files.

IE `compliance-auditor execute --k8s k8s-component.yaml --aws aws-component.yaml --api api/*-component.yaml`

### Why this solution vs policies in a policy engine?
This solution will provide a mechanism for "inheriting" controls that are produced and versioned controlled in tight coordination with the specific tools that are orchestrated. Policies are highly opinionated in this regard to a specific environment and orchestration, and introducing changes in tools that previosuly supported control-satisfaction would require refactoring the policy. 

This allows a report (or System Security Plan) to be generated with potential for overlapping control-satisfaction - without being dependent on a single tool.