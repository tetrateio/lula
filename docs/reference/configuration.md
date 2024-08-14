# Config File

Lula allows the use and specification of a config file in the following ways:
- Checking current working directory for a `lula-config.yaml` file
- Specification with environment variable `LULA_CONFIG=<path>`

## Identification

If identified, Lula will log which configuration file is used to stdout:
```bash
Using config file /home/dev/work/lula/lula-config.yaml
```

## Variables & Constants

Variables and Constants are available for use with validations and OSCAL for the ability to template fields. 

This is executed under the hood by applying a template to a set of data that is established by passing variables into Lula and referencing them in the validation or OSCAL accordingly. 

lula-config.yaml:
```yaml
variables:
  istio:
    healthcheck:
      resource: deployments
      namespace: istio-system
```

validation.yaml
```yaml
metadata:
  name: istio-health-check
  uuid: 67456ae8-4505-4c93-b341-d977d90cb125
domain:
  type: kubernetes
  kubernetes-spec:
    resources:
    - name: istioddeployment
      resource-rule:
        group: apps
        name: istiod
        namespaces:
        - {{ .variables.istio.healthcheck.namespace }}
        resource: {{ .variables.istio.heatlhcheck.resource }}
        version: v1
```
