# 8. Support for Configuration Data

Date: 2024-07-22

## Status

Proposed

## Context

There is an identified need to pull in extra information into various artifacts that Lula operates on. For example
* Adding metadata information to OSCAL documents during Lula Generation methods
* Adding variables and/or secrets into Lula Validation manifests
* Templating component definitions with version numbers

Each of these possible operations may require a distinct input `configuration` file that provides the data, however it is desirable that the underlying libraries and methods support all these possible use cases.

### Details on use cases

#### Build-time template of constants

Addition of system-specific values that are not secrets, but just items that might be changing between system design iterations.

Templating should happen at build-time, i.e., when the component-definition is `composed`

Example: Lula Validation templating namespaces, names, rego values:
```yaml
metadata:
  name: istio-metrics-logging-configured
  uuid: 70d99754-2918-400c-ac9a-319f874fff90
domain:
  type: kubernetes
  kubernetes-spec:
    resources:
    - name: istioConfig
      resource-rule:
        resource: configmaps
        namespaces:
        - "###LULA_CONST_ISTIO_NAMESPACE###"
        version: v1
        name: "###LULA_CONST_ISTIO_CONFIG_NAME###"
        field:
          jsonpath: .data.mesh
          type: yaml
provider:
  type: opa
  opa-spec:
    rego: |
      package validate
      import rego.v1

      # Default values
      default validate := false
      default msg := "Not evaluated"

      # Validate Istio configuration for metrics logging support
      validate if {
        check_metrics_enabled.result
      }
      msg = check_metrics_enabled.msg

      check_metrics_enabled = { "result": false, "msg": msg } if {
        input.istioConfig.###LULA_CONST_ISTIO_PROMETHEUS_MERGE### == false
        msg := "Metrics logging not supported."
      } else = { "result": true, "msg": msg } if {
        msg := "Metrics logging supported."
      }
    output:
      validation: validate.validate
      observations:
      - validate.msg
```

#### Run-time template of variables

Addition of deployment-specific variables that are subject to change across deployments of a system. 

Templating should happen at run-time since these values are likely dependant on the environment and/or possibly an output of some other process therein.

Example: Creating a custom user/API key for a given application, e.g., Keycloak

```bash
#!/bin/bash

# requires https://stedolan.github.io/jq/download/

# config
KEYCLOAK_URL=http://localhost:8080/auth
KEYCLOAK_REALM=realm
KEYCLOAK_CLIENT_ID=clientId
KEYCLOAK_CLIENT_SECRET=clientSecret
USER_ID=userId

export TKN=$(curl -X POST "${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token" \
 -H "Content-Type: application/x-www-form-urlencoded" \
 -d "username=${KEYCLOAK_CLIENT_ID}" \
 -d "password=${KEYCLOAK_CLIENT_SECRET}" \
 -d 'grant_type=password' \
 -d 'client_id=admin-cli' | jq -r '.access_token')

curl -X GET "${KEYCLOAK_URL}/admin/realms/${KEYCLOAK_REALM}/users/${USER_ID}" \
-H "Accept: application/json" \
-H "Authorization: Bearer $TKN" | jq .
```

#### Lula OSCAL Generation

Need to add additional information to OSCAL documents, in this use case specifically looking at SSP generation.

We have some metadata which is constant/managed external to Lula:
```yaml
metadata:
  title: "System Security Plan for UDS Core"
  last-modified: 2024-07-22Z12:00:00
  oscal-version: 1.1.2
```

We have some auto-generated content, e.g., created from the component-definition model
```yaml
system-security-plan:
  system-characteristics:
    system-name: "UDS Core"
  system-implementation:
    components:
      - uuid: f2b245ea-f149-45cf-a740-86081fdb2922
        title: Istio
      - uuid: d1ce0ed3-d678-4bf0-b9f4-330bacd97473
        title: Grafana
```

## (Proposed) Decision

Two separate Lula Config files for the OSCAL (`lula gen`) use cases vs. Validation configuration (`lula validate`) use cases

### Lula Generation

Define a configuration file that when provided to a `generate` command will inject some data into the specified OscalModelSchema jsonpath:

```yaml
kind: LulaOscalConfig

# Map substitutions
maps:
  - name: ssp-metadata
    oscal-key: system-security-plan.metadata
    file: ./metadata.yaml # file OR content specified
    content: |
      metadata:
        title: "System Security Plan for UDS Core"
        last-modified: 2024-07-22Z12:00:00
        oscal-version: 1.1.2
```

Underlying libraries/implementation will be the k8s.io jsonpath module and map merge functions to identify the oscal-key from the OscalModelSchema and inject the contents of `file` or `content` into the schema. Ideally this will manifest as such:

```bash
lula gen ssp --config config-file.yaml --component component-defintion.yaml
```
where the gen ssp uses the component-definition to generate the auto-portions and reads from the map substitutions in `LulaOscalConfig` to inject relevant data where specified.

### Variable substitution

Define a configuration file that when provided to a `validate` (or `compose` or `assess`) command will configure a viper engine to substitute the data:

```yaml
kind: LulaValidationConfig

# Constants substition - gets subbed during composition
constants:
  - name: ...
    value: ...

# Variables - gets subbed at runtime.. from env vars?
variables:
  - name: ...
    value: ...
```

The `constants` are subbed at build-time, whereas the `variables` are subbed during the runtime processes.

## Consequences