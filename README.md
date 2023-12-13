![lint](https://github.com/fdo-rs/fdo-operator/actions/workflows/lint.yaml/badge.svg)
![tests](https://github.com/fdo-rs/fdo-operator/actions/workflows/test.yaml/badge.svg)
[![go report](https://goreportcard.com/badge/github.com/fdo-rs/fdo-operator)](https://goreportcard.com/report/github.com/fdo-rs/fdo-operator)
![Build and push images](https://github.com/fdo-rs/fdo-operator/actions/workflows/images.yaml/badge.svg)

# FDO Operator
The FDO Operator deploys [FIDO Device Onboard (FDO)](https://fidoalliance.org/intro-to-fido-device-onboard/) servers on Red Hat OpenShift.

## Description
The FDO Operator makes it easier to deploy and run any of the FDO servers (manufacturing, rendezvous, or owner-onboarding) on Red Hat OpenShift, catering to both device manufacturers and device owners. It is based on the [open source Rust implementation of FDO](https://github.com/fdo-rs/fido-device-onboard-rs/).

## TODO

Keep in mind that the operator is a work in progress, is highly opinionated and currently has many limitations.

* The owner-onboarding and service-info API servers are deployed as a single unit called the Onboarding server. All communication between the owner-onboarding and the service-info is only within a pod.

* The servers are exposed as OpenShift routes with default generated host names, and support only HTTP on port 80. We intend to allow custom host names, and will consider enabling other protocols if needed.

* The number of replicas is always one, it is currently not possible to scale the deployment.

* The API validation is limited and needs to be updated (e.g. Optional/Requires, default values), as well as the API documentation. Admission webhooks should be added for complex cross-field validations.

* The log level inside FDO containers is TRACE by default and currently cannot be changed.

* Support multiple versions of the FDO server implementation for compatibility reasons, e.g. by maintaining multiple versions of the operator.

* It is not possible to explicitly specify container resources (requests/limits). This should be changed in the future.

* There are currently no liveness or readiness probes.

* There is no place for additional service info configuration in the Onboarding Server CRD. In general, only a limited set of FDO configuration parameters is exposed via the CRDs.

* The names of required secrets (for keys and certificates) and persistent volume claims (for ownership vouchers) are hard-coded. We should allow customizing those, and/or make them include a CR instance name for deduplication within the same namespace.

* Device-specific service-info configuration is currently not supported. Enabling this functionality would require a persistent volume, exposing the admin API via an endpoint, and managing a secret for the admin authentication token.

* Currently, service-info files are automatically added to the onboarding configuration by creating and annotating `ConfigMaps`. Those have size limitations and we may consider other mechanisms as a source of service-info files. In addition, `Secrets` should be supported as a source of sensitive files.

* There is also room for many optimizations and code improvements:

  * Modify the watchers (`Owns()`) to be more selective and watch only relevant resources.
  * Generate a new `ConfigMap` with a random suffix every time the configuration changes to automatically trigger deployment updates.
  * Write a lot more unit tests.
  * Refactor the code for DRY.
  * Remove the use of _github.com/redhat-cop/operator-utils_ as it is outdated.
  * Implement smarter re-queues in case of success and errors in the reconcile logic.
  * Update a resource only if its related part changes instead of trying to do it on every reconciliation attempt.
  * Populate the `Status` of a custom resource to better reflect its state.
  * We can store public certificates in `ConfigMaps` instead of `Secrets` (as usually done in OpenShift).

* Finally, there are a few open questions:

  * How can we make it easier for a user to work with (create, attach) the required persistent volumes?
  * How can we enforce the mandatory secrets (keys, certificates), and respond to any changes in them?

## FDO Server Images

* The operator uses stable [development FDO images](https://quay.io/organization/fido-fdo) by default, although they may not be of the latest version.

* The server CRDs allow changing server images. See the files in [hack/samples](hack/samples/) for examples.

* Make sure to use a server version that is compatible with your FDO client.

* Keep in mind that we currently do not maintain multiple operator versions, therefore cutting edge or too old FDO server images may not supported (e.g. because of incompatible configuration files).

## Getting Started
You will need an OpenShift cluster to run against. You can use [Red Hat OpenShift Local](https://developers.redhat.com/products/openshift-local/overview) to get a local cluster for testing, or run against a remote cluster.

Before some of the custom resources created by the operator can start, they require the following pre-configured Kubernetes resources:

* Keys and certificates as dictated by the FDO implementation. Sample keys and certificates _**for testing**_ can be generated by running

  ```console
  make keys-gen
  ```

  and deployed to the cluster with

  ```console
  make keys-push
  ```

* Persistent volume claims for ownership vouchers. A manufacturing server and an onboarding server both expect a `fdo-ownership-vouchers-pvc`. The volume can be shared if the servers are deployed into the same namespace, making the synchronizing of ownership vouchers automatic (no manual copying will be required in this case).

  **Note:** If you are trying the sample manifests (below) on Red Hat OpenShift Local (CRC), a sample PVC definition is already included and you do not need to create a PVC separately.

To make it easier for a user to manage service info files that will be copied to an onboarded device by FDO, they are stored in `ConfigMaps`. The service-info configuration file is updated accordingly and does not require a user action.

In order to add a file to the service-info, create a `ConfigMap` labeled and annotated as follows, either before or after creating an instance of `FDOOnboardingServer`. In the latter case, the server will be updated to pick up the new file.

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  labels:
    fdo.serviceinfo.file/owner: <onboarding-server-instance>
  annotations:
    fdo.serviceinfo.file/name: <filename>
    fdo.serviceinfo.file/path: /<destination-path>/<destination-filename>
    fdo.serviceinfo.file/permissions: <permissions> # optional, e.g. '755'
  name: <configmap-name>
immutable: false/true
binaryData:
  <filename>: <file-contents>
```

# Sample Deployment

**Note:** This guide assumes that you are running on Red Hat OpenShift Local (CRC) and your current namespace for testing is named `fdo`.

1. Install the operator in [any standard way](https://docs.openshift.com/container-platform/4.12/operators/operator_sdk/golang/osdk-golang-tutorial.html#osdk-run-operator_osdk-golang-tutorial) for operators, or from a catalog image at `ghcr.io/fdo-rs/fdo-operator-catalog`:

   ```console
   oc apply -f hack/openshift/fdo_catalogsource.yaml
   oc apply -f hack/openshift/fdo_operator.yaml
   ```

2. Create the required secrets as described in [Getting Started](#getting-started).

3. Create sample instances and configuration:

  ```console
  oc apply -f hack/samples/
  ```

  The manufacturing server is now available at `http://manufacturing-server-fdo.apps-crc.testing:80`.

You can list generated ownership vouchers by running `exec` in a manufacturing server pod, e.g.

```console
oc exec -ti manufacturing-server-<pod-id> -- ls -1 /etc/fdo/ownership_vouchers
```

And copy an ownership voucher from a pod by running

```console
oc cp manufacturing-server-<pod-id>:/etc/fdo/ownership_vouchers/<device-guid> <device-guid>
```

When testing FDO onboarding using OpenShift Local, you may need to enable traffic between a device and the OpenShift cluster. For instance, if you are simulating a device using a VM, you can allow the VM to access the OpenShift Local (CRC) cluster as explained in [Libvirt routing between two NAT networks](https://serverfault.com/questions/1109903/libvirt-routing-between-two-nat-networks):

```console
sudo iptables -t nat -I POSTROUTING 1 -s 192.168.130.0/24 -d 192.168.124.0/24 -j ACCEPT
sudo iptables -t nat -I POSTROUTING 1 -s 192.168.124.0/24 -d 192.168.130.0/24 -j ACCEPT

sudo iptables -I FORWARD 1 -s 192.168.124.0/24 -d 192.168.130.0/24 -j ACCEPT
sudo iptables -I FORWARD 1 -s 192.168.130.0/24 -d 192.168.124.0/24 -j ACCEPT
```

where `192.168.130.0/24` and `192.168.124.0/24` are the two libvirt networks, one is for CRC (usually `crc`) and the other for VMs (e.g. `default`).

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
