# CDAP Operator

## Overview

This CDAP operator is a custom Kubernetes operator that makes it easy to CDAP on Kubernetes.


### Prerequisites

Ensure that you have satisfied all of the prerequisites of the [operator-framework](https://github.com/operator-framework/operator-sdk#prerequisites).

### Build the operator

Build the CDAP operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=gcr.io/cdap-alpha-dev/cdap-operator:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

### Using the operator

```
# Setup Service Account
$ kubectl create -f deploy/service_account.yaml
# Setup RBAC
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
# Setup the CRD
$ kubectl create -f deploy/crds/io_v1alpha1_cdap_crd.yaml
# Deploy the app-operator
$ kubectl create -f deploy/operator.yaml

# Create an CDAP CR
# The default controller will watch for CDAP objects and create a pod for each CR
$ kubectl create -f deploy/crds/io_v1alpha1_cdap_cr.yaml

# Verify that pods are created
$ kubectl get pods
NAME                            READY   STATUS    RESTARTS   AGE
cdap-instance1-0                1/1     Running   0          10m
cdap-operator-7c6b8dff5-qfht7   1/1     Running   0          10m


# Cleanup
$ kubectl delete -f deploy/crds/io_v1alpha1_cdap_cr.yaml
$ kubectl delete -f deploy/operator.yaml
$ kubectl delete -f deploy/role.yaml
$ kubectl delete -f deploy/role_binding.yaml
$ kubectl delete -f deploy/service_account.yaml
$ kubectl delete -f deploy/crds/io_v1alpha1_cdap_crd.yaml
```
