![lint](https://github.com/empovit/fdo-operator/actions/workflows/lint.yaml/badge.svg)
![tests](https://github.com/empovit/fdo-operator/actions/workflows/test.yaml/badge.svg)
[![codecov](https://codecov.io/gh/empovit/fdo-operator/branch/main/graph/badge.svg?token=EMH9QLP6NR)](https://codecov.io/gh/empovit/fdo-operator)
[![go report](https://goreportcard.com/badge/github.com/empovit/fdo-operator)](https://goreportcard.com/report/github.com/empovit/fdo-operator)
![Build and push images](https://github.com/empovit/fdo-operator/actions/workflows/images.yaml/badge.svg)

# fdo-operator
The FDO Operator deploys [FIDO Device Onboard (FDO)](https://fidoalliance.org/intro-to-fido-device-onboard/) servers on Red Hat OpenShift.

## Description
The FDO Operator makes it easier to deploy and run any of the FDO servers (manufacturing, rendezvous, or owner-onboarding) on Red Hat OpenShift, catering to both device manufacturers and device owners.


https://github.com/fedora-iot/fido-device-onboard-rs/

Before some of the custom resources created by the operator can start, they  require the following pre-configured Kubernetes resources:

* Keys and certificates
* Persistent volume claims (PVCs)



The development is a work in progress and

// TODO(user): An in-depth paragraph about your project and overview of use
* Link to the FDO implementation
* Routes
* Owner+onboarding and service-info API as a single unit (pod)

## Getting Started
You’ll need an OpenShift cluster to run against. You can use [Red Hat OpenShift Local](https://developers.redhat.com/products/openshift-local/overview) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

// TODO:
* Generating keys
* How to create service info files
* Onboarding a VM, enabling traffic between a "device" and the cluster
* PVCs, a shared PV or copying OVs from manufacturer to owner

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/fdo-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/fdo-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

