## Development TODO

* Expose destination namespace in the CRDs
* Update documentation
* Let a user initialize service info files in the PV before the server starts
* Implement unit tests
* Add field validations, Optional/Requires, default values and documentation
* Refactor the code for DRY
* Allow customizing the route hostname
* Should we create/update PVCs or require from user?
* How can we enforce the mandatory secrets?
* How do we reload pods whenever a key/cert secret changes?
* Allow customizing secret names and PVC names
* Implement populating of the Status (esp. Pods, if needed)
* Implement admission webhooks
* Implement additional services info

## Setting Up

1. Generate keys and certificates

   ```sh
   make keys-gen
   ```

2. Create secrets in the namespace where the FDO servers will be deployed:

   ```sh
   make keys-push
   ```

3. Create the following persistent volume claims in the namespace where the FDO servers will be deployed:

   * fdo-ownership-vouchers-pvc
   * fdo-serviceinfo-files-pvc

## Testing Device Initialization

1. Build an `fdo-manufacturing-client` (`fdo-init`) container image:

   ```sh
   cd containers
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