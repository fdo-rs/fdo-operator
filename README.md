## Development with CodeReady Containers (СRC)

* Rendezvous server DNS in the manufacturer configuration is `fdo-rendezvous.apps-crc.testing`
* Image registry is the IP of the local computer (e.g. `192.168.130.1`)
* Using a `hostPath` volume for ownership vouchers

Challenges to address:

* ServiceInfo configuration - initial user, commands, files, encryption, etc.
* ServiceInfo API access token in configuration

## Setting up

1. Generate keys and certificates

   ```sh
   ansible-playbook keys.yml
   ```

2. Create secrets in `fdo-operator`:

   ```sh
   ./secrets.sh
   ```

3. Build images and publish them to a local registry

   ```sh
   ansible-playbook registry.yml
   ```

4. Add the registry to the list of insecure registries in the cluster

   ```sh
   oc edit image.config.openshift.io/cluster
   ```

   and insert

   ```yaml
   spec:
    ...
    registrySources:
        insecureRegistries:
        - 192.168.130.1:5000
   ```

5. Update the registry in the image spec of the pods, e.g. in [manifests/manufacturing.yml](manifests/manufacturing.yml):

   ```yaml
   spec:
     containers:
     - name: fdo-manufacturing-server
       image: 192.168.130.1:5000/fdo-manufacturing-server:latest
   ```

## Testing Device Initialization

1. Build an `fdo-manufacturing-client` (`fdo-init`) container image:

   ```sh
   podman build -t fdo-init-client:latest -f Containerfile.manufacturing-client
   ```

2. Run the client in a container by specifying DIUN configuration and a manufacturing server, e.g.:

   ```sh
   podman run -ti --rm \
      -e DIUN_PUB_KEY_INSECURE=true \
      -e MANUFACTURING_SERVER_URL=http://fdo-manufacturing.apps-crc.testing fdo-init-client:latest
   ```

## Useful commands

* SSH to the CRC VM

  ```sh
  ssh -i ~/.crc/machines/crc/id_ecdsa core@$(crc ip)
  ```

* List generated ownership vouchers

  ```
  oc exec -ti fdo-manufacturing-deployment-<id> -- ls /etc/fdo/ownership_vouchers
  ```

* Copy an ownership voucher from a pod

  ```sh
  oc cp fdo-manufacturing-deployment-<id>:/etc/fdo/ownership_vouchers/<filename> <filename>
  ```

