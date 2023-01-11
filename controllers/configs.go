/*
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
*/

package controllers

import (
	"fmt"

	fdov1 "github.com/empovit/fdo-operators/api/v1"
)

const ServiceInfoAuthToken = "ExampleAuthToken"

type Driver struct {
	Directory *Directory `yaml:"Directory"`
}

type Directory struct {
	Path string `yaml:"path"`
}

func NewDriver(path string) *Driver {
	return &Driver{
		Directory: &Directory{
			Path: path,
		},
	}
}

type OwnerOnboardingServerConfig struct {
	SessionStoreDriver           *Driver                      `yaml:"session_store_driver"`
	OwnerShipVoucherStoreDriver  *Driver                      `yaml:"ownership_voucher_store_driver"`
	Bind                         string                       `yaml:"bind"`
	TrustedDeviceKeysPath        string                       `yaml:"trusted_device_keys_path"`
	OwnerPrivateKeyPath          string                       `yaml:"owner_private_key_path"`
	OwnerPublicKeyPath           string                       `yaml:"owner_public_key_path"`
	OwnerAddresses               []OwnerAddress               `yaml:"owner_addresses"`
	ReportToRendezvousEndpoint   bool                         `yaml:"report_to_rendezvous_endpoint_enabled"`
	ServiceInfoAPIURL            string                       `yaml:"service_info_api_url"`
	ServiceInfoAPIAuthentication ServiceInfoAPIAuthentication `yaml:"service_info_api_authentication"`
}

type OwnerAddress struct {
	Transport string    `yaml:"transport"`
	Port      uint16    `yaml:"port"`
	Addresses []Address `yaml:"addresses"`
}

func NewOwnerAddress(o *fdov1.OwnerAddress) (*OwnerAddress, error) {
	ownAddr := &OwnerAddress{
		Port:      o.Port,
		Transport: o.Transport,
		Addresses: make([]Address, len(o.Addresses)),
	}
	for i, a := range o.Addresses {
		newAddr, err := NewAddress(&a)
		if err != nil {
			return nil, err
		}
		ownAddr.Addresses[i] = *newAddr
	}
	return ownAddr, nil
}

type Address struct {
	DNSName   string `yaml:"dns_name"`
	IPAddress string `yaml:"ip_address"`
}

func NewAddress(o *fdov1.Address) (*Address, error) {
	if o.DNSName != "" && o.IPAddress != "" {
		return nil, fmt.Errorf("cannot use both DNS and IP address for address")
	}
	if o.DNSName == "" && o.IPAddress == "" {
		return nil, fmt.Errorf("either a DNS or IP address is required for rendezvous server")
	}
	return &Address{
		DNSName:   o.DNSName,
		IPAddress: o.IPAddress,
	}, nil
}

type ServiceInfoAPIAuthentication struct {
	BearerToken BearerToken `yaml:"BearerToken"`
}

type BearerToken struct {
	Token string `yaml:"token"`
}

func NewServiceInfoAPIAuthentication(token string) *ServiceInfoAPIAuthentication {
	return &ServiceInfoAPIAuthentication{
		BearerToken: BearerToken{
			Token: token,
		},
	}
}

func (c *OwnerOnboardingServerConfig) setValues(server *fdov1.FDOOnboardingServer) error {
	c.SessionStoreDriver = NewDriver("/etc/fdo/sessions/")
	c.OwnerShipVoucherStoreDriver = NewDriver("/etc/fdo/ownership_vouchers/")
	c.Bind = "0.0.0.0:8081"
	c.TrustedDeviceKeysPath = "/etc/fdo/keys/device_ca_cert.pem"
	c.OwnerPrivateKeyPath = "/etc/fdo/keys/owner_key.der"
	c.OwnerPublicKeyPath = "/etc/fdo/keys/owner_cert.pem"

	if len(server.Spec.OwnerAddresses) == 0 {
		// TODO: Set default if not specified
		return fmt.Errorf("owner addresses must contain at least one value")
	}
	c.OwnerAddresses = make([]OwnerAddress, len(server.Spec.OwnerAddresses))
	for i, a := range server.Spec.OwnerAddresses {
		ownAddr, err := NewOwnerAddress(&a)
		if err != nil {
			return err
		}
		c.OwnerAddresses[i] = *ownAddr
	}
	c.ReportToRendezvousEndpoint = true
	c.ServiceInfoAPIURL = "http://127.0.0.1:8083/device_info"
	c.ServiceInfoAPIAuthentication = *NewServiceInfoAPIAuthentication(ServiceInfoAuthToken)
	return nil
}

type ServiceInfoAPIServerConfig struct {
	Bind                      string      `yaml:"bind"`
	DeviceSpecificStoreDriver *Driver     `yaml:"device_specific_store_driver"`
	ServiceInfoAuthToken      string      `yaml:"service_info_auth_token"`
	ServiceInfo               ServiceInfo `yaml:"service_info"`
}

type ServiceInfo struct {
	InitialUser InitialUser `yaml:"initial_user"`
	// TODO: add commands, files, and disk operations
}

type InitialUser struct {
	Username string   `yaml:"username"`
	SSHKeys  []string `yaml:"sshkeys"`
}

func (c *ServiceInfoAPIServerConfig) setValues(server *fdov1.FDOOnboardingServer) error {
	c.Bind = "0.0.0.0:8083"
	c.DeviceSpecificStoreDriver = NewDriver("/etc/fdo/device_specific_serviceinfo")
	c.ServiceInfoAuthToken = ServiceInfoAuthToken
	// TODO: Update the CRD, validate its service info and copy from it
	return nil
}
