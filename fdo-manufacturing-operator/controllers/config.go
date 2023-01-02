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

type Config struct {
	SessionStoreDriver          Driver           `yaml:"session_store_driver"`
	OwnerShipVoucherStoreDriver Driver           `yaml:"ownership_voucher_store_driver"`
	PublicKeyStoreDriver        Driver           `yaml:"public_key_store_driver"`
	Bind                        string           `yaml:"bind"`
	RendezvousInfo              []RendezvousInfo `yaml:"rendezvous_info"`
	Protocols                   Protocols        `yaml:"protocols"`
	Manufacturing               Manufacturing    `yaml:"manufacturing"`
}

type Driver struct {
	Directory Directory `yaml:"Directory"`
}

type Directory struct {
	Path string `yaml:"path"`
}

type Protocols struct {
	PlainDI bool `yaml:"plain_di"`
	DIUN    DIUN `yaml:"diun,omitempty"`
}

type DIUN struct {
	KeyPath  string `yaml:"key_path"`
	CertPath string `yaml:"cert_path"`
	// Allowed values: SECP256R1 or SECP384R1
	KeyType       string `yaml:"key_type"`
	MFGStringType string `yaml:"mfg_string_type"`
	// Allowed values: FileSystem, Tpm
	AllowedKeyStorageTypes []string `yaml:"allowed_key_storage_types"`
}

type Manufacturing struct {
	ManufacturerCertPath   string `yaml:"manufacturer_cert_path"`
	ManufacturerPrivateKey string `yaml:"manufacturer_private_key"`
	OwnerCertPath          string `yaml:"owner_cert_path"`
	DeviceCertCAPrivateKey string `yaml:"device_cert_ca_private_key"`
	DeviceCertCAChain      string `yaml:"device_cert_ca_chain"`
}

type RendezvousInfo struct {
	DNS        string `yaml:"dns,omitempty"`
	IPAddress  string `yaml:"ipaddress,omitempty"`
	DevicePort uint16 `yaml:"device_port,omitempty"`
	OwnerPort  uint16 `yaml:"owner_port,omitempty"`
	Protocol   string `yaml:"protocol,omitempty"`
}

func setValues(c *Config, fdoServer *fdov1.FDOManufacturing) error {
	c.SessionStoreDriver = *NewDriver("/etc/fdo/sessions/")
	c.OwnerShipVoucherStoreDriver = *NewDriver("/etc/fdo/ownership_vouchers/")
	c.PublicKeyStoreDriver = *NewDriver("/etc/fdo/keys/")
	c.Bind = "0.0.0.0:8080"
	if err := setRendezvousValues(c, fdoServer.Spec.RendezvousServers); err != nil {
		return err
	}

	if err := setProtocolsValues(c, &fdoServer.Spec.Protocols); err != nil {
		return err
	}

	c.Manufacturing = Manufacturing{
		ManufacturerCertPath:   "/etc/fdo/keys/manufacturer_cert.pem",
		ManufacturerPrivateKey: "/etc/fdo/keys/manufacturer_key.der",
		OwnerCertPath:          "/etc/fdo/keys/owner_cert.pem",
		DeviceCertCAPrivateKey: "/etc/fdo/keys/device_ca_key.der",
		DeviceCertCAChain:      "/etc/fdo/keys/device_ca_cert.pem",
	}
	return nil
}

func setRendezvousValues(c *Config, r []fdov1.RendezvousServer) error {
	rendezvousInfo := make([]RendezvousInfo, len(r), 0)
	if len(r) == 0 {
		return fmt.Errorf("rendezvous servers must contain at least one value")
	}
	for i, s := range r {
		if s.DNS != "" && s.IPAddress != "" {
			return fmt.Errorf("cannot use both DNS and IP address for rendezvous server")
		}

		if s.DNS == "" && s.IPAddress == "" {
			return fmt.Errorf("either a DNS or IP address is required for rendezvous server")
		}
		rendezvousInfo[i] = RendezvousInfo(s)
	}
	c.RendezvousInfo = rendezvousInfo
	return nil
}

func setProtocolsValues(c *Config, p *fdov1.Protocols) error {
	c.Protocols = Protocols{}
	return nil
}

func NewDriver(path string) *Driver {
	return &Driver{
		Directory: Directory{
			Path: path,
		},
	}
}
