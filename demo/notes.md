# Notes for Kyverno querying capabilities

## Prerequisites

- A running kubernetes cluster
- Istio deployed into the cluster so that we can create PeerAuthentication resources (using Big Bang for this)
- A compiled `lula` binary in the root of the repository

## Configuration and Testing

`oscal-component.yaml` defines that each specified namespace should have at least one PeerAuthentication resource present.

The test namespaces being used:

- foo
- test
- test1
- test2

`peer-auth.yaml` has been added to define PeerAuthentication resources to be deployed to our test namespaces. When this file is applied to the cluster, it will create one PeerAuthentication resource in each test namespace.

To create the test namespaces and execute validation with lula:

```bash
./demo/fail.sh
```

The output shows 4 resources (namespaces in this case) failing validation as we expect:

```bash
namespace/foo created
namespace/test created
namespace/test1 created
namespace/test2 created

Applying 1 policy rule to 12 resources...

policy 42c2ffdc-5f05-44df-a67f-eec8660aeffd -> resource /Namespace/test failed: 
1. istio-controlplane_AC-4(peer-auth-for-every-namespace): The specified namespaces must have at least 1 PeerAuthentication resource. 

policy 42c2ffdc-5f05-44df-a67f-eec8660aeffd -> resource /Namespace/test2 failed: 
1. istio-controlplane_AC-4(peer-auth-for-every-namespace): The specified namespaces must have at least 1 PeerAuthentication resource. 

policy 42c2ffdc-5f05-44df-a67f-eec8660aeffd -> resource /Namespace/foo failed: 
1. istio-controlplane_AC-4(peer-auth-for-every-namespace): The specified namespaces must have at least 1 PeerAuthentication resource. 

policy 42c2ffdc-5f05-44df-a67f-eec8660aeffd -> resource /Namespace/test1 failed: 
1. istio-controlplane_AC-4(peer-auth-for-every-namespace): The specified namespaces must have at least 1 PeerAuthentication resource. 
UUID: 42C2FFDC-5F05-44DF-A67F-EEC8660AEFFD
        Resources Passing: 0
        Resources Failing: 4
        Status: Fail
```

Now let's create a PeerAuthentication resource for each test namespace and re-run validation with lula:

```bash
./demo/pass.sh
```

The output shows 4 passing resources now:

```bash
peerauthentication.security.istio.io/lula-test created
peerauthentication.security.istio.io/lula-test created
peerauthentication.security.istio.io/lula-test created
peerauthentication.security.istio.io/lula-test created

Applying 1 policy rule to 12 resources...
UUID: 42C2FFDC-5F05-44DF-A67F-EEC8660AEFFD
        Resources Passing: 4
        Resources Failing: 0
        Status: Pass
```

Delete the test namespaces:

```bash
./demo/cleanup.sh
```

Rinse and repeat 😎
