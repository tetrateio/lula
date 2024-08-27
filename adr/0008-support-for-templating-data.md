# 8. Support for Templating Data

Date: 2024-08-27

## Status

Proposed

## Context

There is an identified need to pull in extra information into various artifacts that Lula operates on. The following are the current layers of exploration for templating data:
* Adding variables and/or secrets into runtime processing (IE available for use in templating operations)
* Templating provided artifacts - both in isolation and during lula workflows

Each of these operations fall within the purview of "templating" information and may require various input `configuration` values to provide the data, and additionally may require different workflows under the hood to support the implementations. This document is an exploration of the use cases and identified workflows needed to support.

## Variable Definition

This encompasses the structure of the configuration data as it exists between:
- configuration file
- command line arguments
- environment variables

Lula will require the ability to delineate between sensitive and non-sensitive data to be templated through any of the above methods. 

When precedence has been established between the above methods, having the data available for processing allows for templating to handle how data is applied against a given artifact.

## Templating

Templating support for Lula requires the following:
- Generic go templating support
- Ability to establish a key as sensitive
  - This value should only templated during runtime execution
- (Proposed) Ability to support pre-determined templating functions

## Constraints

Lula will require valid json/yaml during operations which required marshalling data to the applicable model and validation data types. 

Workflows for complex templating will be enabled through a (proposed) `template` command. Enabling users to use templating functions with an outcome that is valid json/yaml OSCAL artifact prior to use in capabilities that require schema compliant artifacts. 

Any templating field used in transient OSCAL contexts will require valid use of json/yaml formatting. This typically will entail string substitution and more simple variable use scenarios. 

### Detailed Use Case Exploration

1. User wants to have composed OSCAL model, but keep configuration separate, templated at run-time

Requires runtime flag being present to signal performing compose without templating.

2. User wants to have composed OSCAL model, WITH validations templated at compose time 
  
Requires runtime flag being present to signal performing compose with templating. A `lula-config.yaml` is required for this operation.

>[!NOTE] I think you do want to support both - the first use case is probably where the "composed" file is taken to different locations, where you don't want to have to port around all the individual validations but may need to operate on them and want the config still broken out. The second use case is more relevant to the "composed" file being an artifact of the `lula validate` command, where you'd want that file to possibly evaluate after the run (thinking in CI scenario)
>
> If we are going to support native go templating and expose the built-in functions (and maybe future additions of our own) then we will need to handle both identification of items that should be templated during compose as well as other items (typically sensitive) that should always remain un-templated. Both scenarios will require delineation between sensitive and non-sensitive variables/templates. 
> 
> Options:
> 1. Perform a pre-template find of "sensitive" items -> replace with a unique item that retains the original path but will not be templated -> template the file -> replace the unique identifiers with the original template fields
>   1.a. This isn't overly difficult to support. For the limited operations that write information to a persistent file we can perform a regex replace before and after templating to retain secret templates.

3. Environment variables are going to be templated values.

Establish order of precedence for the use of environment variables and allow for this data to merge with existing configuration and command-line data prior to templating operations. 

>[!NOTE] This suffers from the same problem as variables generically - we need the ability to identify a variable as sensitive or non-sensitive. 

4. Templating Sensitive Keys

This is a constraint that is present in all underlying variable use currently. Need to delineate between sensitive keys and non-sensitive keys in a template.

5. User wants to have a templated OSCAL model, e.g., some links may be templated if root path changes

6. Validation configuration values can be --set at the command line, e.g., `lula validate -f ./component-definition.yaml --set .some-value=abc123`

#### Use Cases 1 & 2: Sample Lula Validation and associated config that could be templated at build-time

Addition of system-specific values that are not secrets, but just items that might be changing between system design iterations.

Validation with template values, currently structured in go-template syntax:
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
        - "{{ .istio.namespace }}"
        version: v1
        name: "{{ .istio.config-name }}"
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
        input.istioConfig.{{ .istio.prometheus-merge }} == false
        msg := "Metrics logging not supported."
      } else = { "result": true, "msg": msg } if {
        msg := "Metrics logging supported."
      }
    output:
      validation: validate.validate
      observations:
      - validate.msg
```

Validation configuration file (variable schema, based on how it's called in the go-template syntax):
```yaml
istio:
  namespace: istio-system
  config-name: istio-config
  prometheus-merge: enablePrometheusMerge
```
^^ This is assuming we are using go-template under the hood. A different, more rigid structure might be needed if a different templating method is used.

With a component definition that links the above validation, run:
```shell
lula t compose -f ./component-definition.yaml --config ./my-config.yaml
```

#### Use Case 3 & 4: Run-time template of variables

Addition of deployment-specific variables that are subject to change across deployments of a system. 

Templating should happen at run-time since these values are likely dependant on the environment and/or possibly an output of some other process therein.

Example: Creating a custom host/API token for a given application, e.g., Keycloak
```yaml
metadata:
  name: check-keycloak-api
  uuid: bf0aeb97-6e37-4bf4-b976-5f6af6fa81a3
domain:
  type: api
  api-spec:
    name: keycloakAdmin
    endpoint: http://{{ .var.KEYCLOAK_HOST }}:8080/auth/admin/realms/master
    method: GET
    headers:
      Authorization: Bearer {{ .secret.KEYCLOAK_TOKEN }}
provider:
  type: opa
  opa-spec:
    rego: |
      package validate
      import rego.v1

      # Some validation logic here...
```

Here, no config is required, the `var` and `secret` values are provided by the environment. The .var values are persisted in the OSCAL artifact, while the .secret values are not.


#### Use Case 5: Template of OSCAL data

Aside from templating entire sections of OSCAL, maybe some overrides with respect to a linked path might be needed:

```yaml
component-definition:
  components:
    - # ... list of components
      control-implementations:
        - # ... list of control implementations
          implemented-requirements:
            - # ... list of implemented requirements
              links:
                - href: '{{ .rootUrl }}/validations/istio/healthcheck/validation.yaml'
                  rel: lula
                  text: Check that Istio is healthy
```

## Templating Operation

### Proposed Decision 1 - .tpl extensions

Use of Viper with further specification of the `lula-config.yaml` file.

Propose templated files use .tpl/.tmpl extensions to indicate that they are templated / allow for us to handle them differently.
-> Only use .tpl/.tmpl files in validations, oscal model files should remain valid yaml/oscal
-> If you are templating an OSCAL file, must still be valid yaml e.g., 
```yaml
# ... component-definition
  links:
    - href: '{{ const.rootUrl }}/validations/istio/healthcheck/validation.yaml'
      rel: lula
      text: Check that Istio is healthy
```


#### Consequences

Additional features to develop:
- Update to composition to support composing template files + possibly adding a prop or something to indicate that the file is a tmpl vs yaml
- Add templating routines that gets implemented in `validate`, `compose`, `dev`
  -> Multiple layers of templating -
    - Template All -> Templated during `dev`
    - Template only Constants & Variables -> Can be templated during `compose`
    - Template Secrets (presumably const and vars are templated) -> Templated during `validate`, presumable after `compose`+template so that the oscal artifact is purely in memory and the compose+template version can be output as an artifact 
- Add --template flag to `lula compose` to direct the composition to template out the validations and component-definition
- Add --set value support (might be a viper thing)

### Proposed Decision 2 - Most Permissive / Least Controlled

Given the expectations that a provided OSCAL artifact must be schema compliant and valid json/yaml, proposing Lula provide tooling and capabilities that enable the use of templating in a permissive manner that doesn't tightly controlled use.

Therefor enabling the users of this templating feature to craft artifacts with as little or as much of the built-in complexity of go templating functionality while validating for correctness and schema compliance where required and returning concise errors in the event of any issue. 

This will still require methods for identification of sensitive variables 

#### Consequences

- Creation of a `lula tools template` command for a workflow where an end user wants to incorporate complex templating logic before executing a process that expects valid json/yaml. 
- Error handling required for determination of malformed data. Could be natively reported from the `unmarshall` operation

## Variable Structure

### Decision 1 - const / var / secret 

Proposed `lula-config.yaml` structure:
```yaml
log_level : 'debug'

constants: # map[string]interface{}, can be any structured data -> rendered in oscal -> referenced as {{ const.istio.namespace}}
  istio:
    namespace: istio-system
    config:
      name: istio-config
      prometheus-merge: enablePrometheusMerge

variables: # map[string]string, represents environment variables -> rendered in oscal -> referenced as {{ var.some_env_var }}
  some_env_var: some_value
  another_env_var: another_value

secrets: # map[string]string, represents secrets -> NOT rendered in oscal -> referenced as {{ secret.some_secret }}, also pulled from the environment(?) or should we think about k8s secret or other support? -> I think maybe if we do this it should be scoped to a domain
  some_secret: some_value
  another_secret: another_value

```

Intent between defining "constants" vs "variables" vs "secrets" is that constants are not expected to change, variables are expected to change/are environment variables, and secrets are expected to be sensitive and not rendered in the OSCAL artifact.

### Decision 2 - var / sensitive

Proposed `lula-config.yaml` structure:
```yaml
log_level : 'debug'

variables: # map[string]interface{}
  some_env_var: 
    nested_var: some_nested_value
  another_env_var: another_value

secret: # map[string]interface{}
  some_secret:
    some_nested_secret: some_nested_secret_value
  another_secret: another_value

```

Use of delineating `variables` from `secrets` allows for more dynamic templating function use with the ability to identify and retract sensitive data from being available during templating operations. Command flags such as `--variable` and `--secret` (other other names) could be present to set data in these maps from the command line. Environment variables with similar prefixes `LULA_VAR_SOME_ENV_VAR` and `LULA_SECRET_SOME_SECRET` could also be merged into the map prior to templating. 