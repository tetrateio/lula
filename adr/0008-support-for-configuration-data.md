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

#### Template constants, e.g., namespaces in a domain, kubernetes-spec
Happens at build-time, this can be `composed` into the oscal

E.g., templating namespaces, names, rego values to check:
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
        - "###ZARF_CONST_ISTIO_NAMESPACE###"
        version: v1
        name: "###ZARF_CONST_ISTIO_CONFIG_NAME###"
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
        input.istioConfig.###ZARF_CONST_ISTIO_PROMETHEUS_MERGE### == false
        msg := "Metrics logging not supported."
      } else = { "result": true, "msg": msg } if {
        msg := "Metrics logging supported."
      }
    output:
      validation: validate.validate
      observations:
      - validate.msg
```

#### Template secrets, e.g., API Keys
Happens at run-time, this is dependant on the environment, possibly an output of some other process therein...

-> Would you have to create a custom user/API key for a given application, e.g., Keycloak?

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

#### SSP Generation
Some metadata:
```yaml
metadata:
  title: "System Security Plan for UDS Core"
  last-modified: 2024-07-22Z12:00:00
  oscal-version: 1.1.2
```

(auto-generated from component-definition)
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

`lula gen ssp -f <comma-separated-list-of-other-info-yaml> -f <component.yaml>`
-> This feels probably a bit more unique than a simple template of variables/secrets since you're merging entire maps

I don't think this workflow is going to work because if you need to properly define an entire oscal model then you'll be overwriting data when doing the merge -_-

### Options
**Go Templating engine**
This is manifested as {{ .whatever }} within the yaml. 
Pro: It is versatile in that you can define a configuration yaml into any structure you want, and just reference that path in your template. 
Con: Doesn't really support secrets or updating the variables.

**Viper**
Library used for configuration of go applications.
Pro: Proven in Zarf
Con: Have to implement our own custom version, which will likely very closely align to Zarf...
Also, probably need to do custom whatever to handle the yaml stuff

**Zarf/UDS CLI**
Can we just leverage their implementation instead of having to create and manage our own?
Pro: we don't have to manage any of that and can leverage the existing tools to do so. basically if we're packaging up UDS Core as effectively a zarf package, we could include these settings as additional variables within that package config, then on "deploy" those get templated out into local?
Con: requires Zarf and/or UDS Bundles to set-up/handle variable and constant injection. not a native Lula feature. requires `deploy` which can't really work without a cluster, so how does this extend to other non-k8s scenarios?
-> How would this work in BB? Would you create like a zarf package for bb then throw in the templated compliance stuff?
-> Also, what is the order of operations for creating a cluster, deploying an app, creating an API key to access app, templating that variable into the component-defn, running Lula validate -> feels like some of this happens after the `deploy` so how would templating work there?

## Decision

```yaml
kind: LulaConfig

# Map substitutions
maps:
  - name: some-map
    parent-path: system-security-plan
    file: ./ssp-metadata.yaml

# Constants substition - gets subbed during composition
constants:
  - name: ...
    value: ...

# Variables - gets subbed at runtime.. from env vars?
variables:
  - name: ...
    value: ...
```

## Consequences