## Development with CodeReady Containers (СRC)

* Rendezvous server DNS in the manufacturer configuration is `fdo-rendezvous.apps-crc.testing`
* Image registry is the IP of the local computer (e.g. `192.168.130.1`)
* Using a `hostPath` volume for ownership vouchers
* Without [additional setup](#connecting-to-crc-cluster), the cluster is reachable only from the local machine

Challenges to address:

* ServiceInfo configuration - initial user, commands, files, encryption, etc.
* ServiceInfo API access token in configuration

## Setting Up

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

## Useful Commands

* SSH to the CRC VM

  ```sh
  ssh -i ~/.crc/machines/crc/id_ecdsa core@$(crc ip)
  ```

* List generated ownership vouchers

  ```
  oc exec -ti fdo-manufacturing-deployment-<id> -- ls -1 /etc/fdo/ownership_vouchers
  ```

* Copy an ownership voucher from a pod

  ```sh
  oc cp fdo-manufacturing-deployment-<id>:/etc/fdo/ownership_vouchers/<filename> <filename>
  ```

* Change the FDO manufacturing server of a simplified installer image

  When booting from a simplified installer ISO, press `e` before installing RHEL, then edit kernel arguments:

  ```console
  ... fdo.manufacturing_server_url=http://hostname:8080 fdo.diun_pub_key_insecure=true ...
  ```

## Connecting to CRC Cluster

In order to connect to a CRC cluster remotely, a proxy or an SSH tunnel (e.g. using `sshuttle`) must be set up.

In order to allow other VMs (e.g. on `default` network) to access a CRC cluster, which is connected to `crc` network, configure the following, as explained in [Libvirt routing between two NAT networks](https://serverfault.com/questions/1109903/libvirt-routing-between-two-nat-networks)

```console
sudo iptables -t nat -I POSTROUTING 1 -s 192.168.130.0/24 -d 192.168.122.0/24 -j ACCEPT
sudo iptables -t nat -I POSTROUTING 1 -s 192.168.122.0/24 -d 192.168.130.0/24 -j ACCEPT

sudo iptables -I FORWARD 1 -s 192.168.122.0/24 -d 192.168.130.0/24 -j ACCEPT
sudo iptables -I FORWARD 1 -s 192.168.130.0/24 -d 192.168.122.0/24 -j ACCEPT
```

where `192.168.130.0/24` and `192.168.122.0/24` are the two libvirt networks.