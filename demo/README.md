# Demo

## Build Lula binary and create cluster

```bash
./demo/setup.sh
```

## Create demo namespace and pod that will fail validation

```bash
./demo/apply_fail.sh
```

## Use Lula to validate non-root configuration in the pod specification

The results should show 1 failing resource

```bash
./demo/validate.sh
```

## Create a pod that will pass validation

```bash
./demo/apply_pass.sh
```

## Use Lula to validate non-root configuration on the pod specification

The results should show 1 failing resource and 1 passing resource

```bash
./demo/validate.sh
```

## Examine the generated compliance reports

```yaml
- result: Fail
  source-requirements:
    control-id: ac-6
    description: Employ the principle of least privilege, allowing only authorized
      accesses for users (or processes acting on behalf of users) that are necessary
      to accomplish assigned organizational tasks.
    rules:
    - exclude:
        resources: {}
      generate:
        clone: {}
        cloneList: {}
      match:
        resources:
          kinds:
          - Pod
          namespaces:
          - demo
      mutate: {}
      name: ac-6
      validate:
        message: Containers running as root user are not allowed in the demo namespace.
        pattern:
          spec:
            containers:
            - securityContext:
                runAsNonRoot: true
    uuid: 42C2FFDC-5F05-44DF-A67F-EEC8660AEFFD
```

## Teardown the cluster

```bash
./demo/cleanup.sh
```
