![lint](https://github.com/empovit/fdo-operator/actions/workflows/lint.yaml/badge.svg)
![tests](https://github.com/empovit/fdo-operator/actions/workflows/test.yaml/badge.svg)
[![codecov](https://codecov.io/gh/empovit/fdo-operator/branch/main/graph/badge.svg?token=EMH9QLP6NR)](https://codecov.io/gh/empovit/fdo-operator)
[![go report](https://goreportcard.com/badge/github.com/empovit/fdo-operator)](https://goreportcard.com/report/github.com/empovit/fdo-operator)
![Build and push images](https://github.com/empovit/fdo-operator/actions/workflows/images.yaml/badge.svg)

# fdo-operator
The FDO Operator deploys [FIDO Device Onboard (FDO)](https://fidoalliance.org/intro-to-fido-device-onboard/) servers on Red Hat OpenShift.

## Description
The FDO Operator makes it easier to deploy and run any of the FDO servers (manufacturing, rendezvous, or owner-onboarding) on Red Hat OpenShift, catering to both device manufacturers and device owners. It is based on the [Fedora IoT implementation of FDO](https://github.com/fedora-iot/fido-device-onboard-rs/).


Keep in mind that the operator is a work in progress, is highly opinionated and currently has many limitations.

* The owner-onboarding and service-info API servers are deployed as a single unit called the Onboarding server. All communication between the owner-onboarding and the service-info is only within a pod.

* The servers are exposed as OpenShift routes with default generated host names, and support only HTTP on port 80. We intend to allow custom host names, and will consider enabling other protocols if needed.

* The number of replicas is always one, it is currently not possible to scale the deployment.

* The API validation is limited and needs to be updated (e.g. Optional/Requires, default values), as well as the API documentation. Admission webhooks should be added for complex cross-field validations.

* The log level inside FDO containers is TRACE by default and currently cannot be changed.

* The container images the operator is used by default are stable but not maintained. It could be better to use either the [development FDO images](https://quay.io/organization/fido-fdo) or Red Hat certified images for FDO once available.

* It is not possible to explicitly specify container resources (requests/limits). This should change.

* There are currently no liveness or readiness probes.

* There is no place for additional service info configuration in the Onboarding Server CRD. In general, only a limited set of FDO configuration parameters is exposed via the CRDs.

* The names of required secrets (for keys and certificates) and persistent volume claims (for ownership vouchers) are hard-coded. We should allow customizing those, and/or make them include a CR instance name for deduplication within the same namespace.

* Device-specific service-info configuration is not supported. Enabling this functionality would require a persistent volume, exposing the admin API via an endpoint, and managing a secret for the admin authentication token.

* Currently, service-info files are automatically added to the onboarding configuration by creating and annotating `ConfigMaps`. Those have size limitations and we may consider other mechanisms as sources of service-info files. In addition, `Secrets` should be supported as a source of sensitive files.

* There is also room for many optimizations and code improvements:

  * Modify the watchers (`Owns()`) to be more selective and watch only relevant resources.
  * Generate a new `ConfigMap` with a random suffix every time the configuration changes to automatically trigger deployment updates.
  * Write a lot more unit tests.
  * Refactor the code for DRY
  * Remove the use of _github.com/redhat-cop/operator-utils_ as outdated
  * Implements smarter re-queues in case of success and errors in the reconcile logic
  * Update an object only if its related part changes instead of trying to do it on every reconciliation attempt
  * Populate the `Status` of a custom resource to better reflect it state
  * We can store public certificates in `ConfigMaps` instead of `Secrets` (as usually done in OpenShift).

* Finally, there are a few open questions:

  * How can we make it easier for a user to work with (create, attach) the required persistent volumes?
  * How can we enforce the mandatory secrets (keys, certificates), and respond to any changes in them?

## Getting Started
You’ll need an OpenShift cluster to run against. You can use [Red Hat OpenShift Local](https://developers.redhat.com/products/openshift-local/overview) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

Before some of the custom resources created by the operator can start, they require the following pre-configured Kubernetes resources:

* Keys and certificates and dictated by the FDO implementation.
* Persistent volume claims for ownership vouchers. A manufacturing server and an onboarding server both expect a `fdo-ownership-vouchers-pvc`. The volume can be shared if the servers are deployed into the same namespace, making the synchronizing of ownership vouchers automatic (no manual copying will be required in this case).

// TODO:
* Generating keys

1. Generate keys and certificates

   ```sh
   make keys-gen
   ```

2. Create secrets in the namespace where the FDO servers will be deployed:

   ```sh
   make keys-push
   ```

* How to create service info files

* Onboarding a VM, enabling traffic between a "device" and the cluster

    In order to connect to a CRC cluster remotely, a proxy or an SSH tunnel (e.g. using `sshuttle`) must be set up.

    In order to allow other VMs (e.g. on `default` network) to access a CRC cluster, which is connected to `crc` network, configure the following, as explained in [Libvirt routing between two NAT networks](https://serverfault.com/questions/1109903/libvirt-routing-between-two-nat-networks)

    ```console
    sudo iptables -t nat -I POSTROUTING 1 -s 192.168.130.0/24 -d 192.168.122.0/24 -j ACCEPT
    sudo iptables -t nat -I POSTROUTING 1 -s 192.168.122.0/24 -d 192.168.130.0/24 -j ACCEPT

    sudo iptables -I FORWARD 1 -s 192.168.122.0/24 -d 192.168.130.0/24 -j ACCEPT
    sudo iptables -I FORWARD 1 -s 192.168.130.0/24 -d 192.168.122.0/24 -j ACCEPT
    ```

    where `192.168.130.0/24` and `192.168.122.0/24` are the two libvirt networks.

* Manually copying an ownership voucher from a manufacturing server to an onboarding server

* List generated ownership vouchers

  ```
  oc exec -ti fdo-manufacturing-deployment-<id> -- ls -1 /etc/fdo/ownership_vouchers
  ```

* Copy an ownership voucher from a pod

  ```sh
  oc cp fdo-manufacturing-deployment-<id>:/etc/fdo/ownership_vouchers/<filename> <filename>
  ```


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