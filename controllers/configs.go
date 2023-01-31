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

	fdov1alpha1 "github.com/empovit/fdo-operator/api/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
)

const ServiceInfoAuthToken = "ExampleAuthToken"

type Driver struct {
	Directory *Directory `yaml:"Directory,omitempty"`
}

type Directory struct {
	Path string `yaml:"path,omitempty"`
}

func NewDriver(path string) *Driver {
	return &Driver{
		Directory: &Directory{
			Path: path,
		},
	}
}

type OwnerOnboardingServerConfig struct {
	SessionStoreDriver           *Driver                       `yaml:"session_store_driver"`
	OwnerShipVoucherStoreDriver  *Driver                       `yaml:"ownership_voucher_store_driver"`
	Bind                         string                        `yaml:"bind"`
	TrustedDeviceKeysPath        string                        `yaml:"trusted_device_keys_path"`
	OwnerPrivateKeyPath          string                        `yaml:"owner_private_key_path"`
	OwnerPublicKeyPath           string                        `yaml:"owner_public_key_path"`
	OwnerAddresses               []OwnerAddress                `yaml:"owner_addresses"`
	ReportToRendezvousEndpoint   bool                          `yaml:"report_to_rendezvous_endpoint_enabled"`
	ServiceInfoAPIURL            string                        `yaml:"service_info_api_url"`
	ServiceInfoAPIAuthentication *ServiceInfoAPIAuthentication `yaml:"service_info_api_authentication"`
}

type OwnerAddress struct {
	Transport string    `yaml:"transport"`
	Port      uint16    `yaml:"port"`
	Addresses []Address `yaml:"addresses"`
}

type Address struct {
	DNSName   string `yaml:"dns_name,omitempty"`
	IPAddress string `yaml:"ip_address,omitempty"`
}

type ServiceInfoAPIAuthentication struct {
	BearerToken *BearerToken `yaml:"BearerToken,omitempty"`
}

type BearerToken struct {
	Token string `yaml:"token,omitempty"`
}

func NewServiceInfoAPIAuthentication(token string) *ServiceInfoAPIAuthentication {
	return &ServiceInfoAPIAuthentication{
		BearerToken: &BearerToken{
			Token: token,
		},
	}
}

func (c *OwnerOnboardingServerConfig) setValues(server *fdov1alpha1.FDOOnboardingServer, route *routev1.Route) error {
	c.SessionStoreDriver = NewDriver("/etc/fdo/sessions/")
	c.OwnerShipVoucherStoreDriver = NewDriver("/etc/fdo/ownership_vouchers/")
	c.Bind = "0.0.0.0:8081"
	c.TrustedDeviceKeysPath = "/etc/fdo/keys/device_ca_cert.pem"
	c.OwnerPrivateKeyPath = "/etc/fdo/keys/owner_key.der"
	c.OwnerPublicKeyPath = "/etc/fdo/keys/owner_cert.pem"

	// For now use a highly opinionated config that points to
	// a generated OpenShift route and uses HTTP transport
	c.OwnerAddresses = []OwnerAddress{
		{
			Transport: "http",
			Port:      80,
			Addresses: []Address{
				{
					DNSName: route.Spec.Host,
				},
			},
		},
	}
	c.ReportToRendezvousEndpoint = true
	c.ServiceInfoAPIURL = "http://127.0.0.1:8083/device_info"
	c.ServiceInfoAPIAuthentication = NewServiceInfoAPIAuthentication(ServiceInfoAuthToken)
	return nil
}

type ServiceInfoAPIServerConfig struct {
	Bind                      string       `yaml:"bind"`
	DeviceSpecificStoreDriver *Driver      `yaml:"device_specific_store_driver"`
	ServiceInfoAuthToken      string       `yaml:"service_info_auth_token"`
	ServiceInfoAdminAuthToken string       `yaml:"admin_auth_token,omitempty"`
	ServiceInfo               *ServiceInfo `yaml:"service_info"`
}

type ServiceInfo struct {
	InitialUser            *ServiceInfoInitialUser           `yaml:"initial_user,omitempty"`
	Files                  []ServiceInfoFile                 `yaml:"files,omitempty"`
	Commands               []ServiceInfoCommand              `yaml:"commands,omitempty"`
	DiskEncryptionClevises []ServiceInfoDiskEncryptionClevis `yaml:"diskencryption_clevis,omitempty"`
}

type ServiceInfoInitialUser struct {
	Username string   `yaml:"username"`
	SSHKeys  []string `yaml:"sshkeys"`
}

type ServiceInfoFile struct {
	Path        string `yaml:"path"`
	Permissions string `yaml:"permissions,omitempty"`
	SourcePath  string `yaml:"source_path"`
	ConfigMap   string `yaml:"-"`
}

type ServiceInfoCommand struct {
	Command      string   `yaml:"command"`
	Args         []string `yaml:"args"`
	MayFail      bool     `yaml:"may_fail"`
	ReturnStdOut bool     `yaml:"return_stdout"`
	ReturnStdErr bool     `yaml:"return_stderr"`
}

type ServiceInfoDiskEncryptionClevis struct {
	DiskLabel string                                  `yaml:"disk_label"`
	Binding   *ServiceInfoDiskEncryptionClevisBinding `yaml:"binding"`
	ReEncrypt bool                                    `yaml:"reencrypt"`
}

type ServiceInfoDiskEncryptionClevisBinding struct {
	Pin    string `yaml:"pin,omitempty"`
	Config string `yaml:"config,omitempty"`
}

func (c *ServiceInfoAPIServerConfig) setValues(server *fdov1alpha1.FDOOnboardingServer, files []ServiceInfoFile) error {
	c.Bind = "0.0.0.0:8083"
	c.DeviceSpecificStoreDriver = NewDriver("/etc/fdo/device_specific_serviceinfo")
	c.ServiceInfoAuthToken = ServiceInfoAuthToken
	c.ServiceInfo = &ServiceInfo{}
	if server.Spec.ServiceInfo == nil {
		return nil
	}
	if server.Spec.ServiceInfo.InitialUser != nil {
		user := server.Spec.ServiceInfo.InitialUser
		c.ServiceInfo.InitialUser = &ServiceInfoInitialUser{
			Username: user.Username,
			SSHKeys:  user.SSHKeys,
		}
	}
	c.ServiceInfo.Files = files
	if server.Spec.ServiceInfo.Commands != nil {
		commands := server.Spec.ServiceInfo.Commands
		c.ServiceInfo.Commands = make([]ServiceInfoCommand, len(commands))
		for i, cmd := range commands {
			c.ServiceInfo.Commands[i] = ServiceInfoCommand(cmd)
		}
	}
	if server.Spec.ServiceInfo.DiskEncryptionClevises != nil {
		clevises := server.Spec.ServiceInfo.DiskEncryptionClevises
		c.ServiceInfo.DiskEncryptionClevises = make([]ServiceInfoDiskEncryptionClevis, len(clevises))
		for i, clv := range clevises {
			c.ServiceInfo.DiskEncryptionClevises[i] = NewServiceInfoDiskEncryptionClevis(clv)
		}
	}
	return nil
}

func NewServiceInfoDiskEncryptionClevis(cl fdov1alpha1.DiskEncryptionClevis) ServiceInfoDiskEncryptionClevis {
	c := ServiceInfoDiskEncryptionClevis{
		DiskLabel: cl.DiskLabel,
		ReEncrypt: cl.ReEncrypt,
	}
	if cl.Binding != nil {
		c.Binding = &ServiceInfoDiskEncryptionClevisBinding{
			Pin:    cl.Binding.Pin,
			Config: cl.Binding.Config,
		}
	}
	return c
}

type ManufacturingServerConfig struct {
	SessionStoreDriver          *Driver          `yaml:"session_store_driver"`
	OwnerShipVoucherStoreDriver *Driver          `yaml:"ownership_voucher_store_driver"`
	PublicKeyStoreDriver        *Driver          `yaml:"public_key_store_driver"`
	Bind                        string           `yaml:"bind"`
	RendezvousInfo              []RendezvousInfo `yaml:"rendezvous_info"`
	Protocols                   *Protocols       `yaml:"protocols"`
	Manufacturing               *Manufacturing   `yaml:"manufacturing"`
}

type Protocols struct {
	PlainDI bool  `yaml:"plain_di"`
	DIUN    *DIUN `yaml:"diun,omitempty"`
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

func (c *ManufacturingServerConfig) setValues(server *fdov1alpha1.FDOManufacturingServer) error {
	c.SessionStoreDriver = NewDriver("/etc/fdo/sessions/")
	c.OwnerShipVoucherStoreDriver = NewDriver("/etc/fdo/ownership_vouchers/")
	c.PublicKeyStoreDriver = NewDriver("/etc/fdo/keys/")
	c.Bind = "0.0.0.0:8080"
	if err := c.setRendezvousValues(server.Spec.RendezvousServers); err != nil {
		return err
	}

	if err := c.setProtocolsValues(server.Spec.Protocols); err != nil {
		return err
	}

	c.Manufacturing = &Manufacturing{
		ManufacturerCertPath:   "/etc/fdo/keys/manufacturer_cert.pem",
		ManufacturerPrivateKey: "/etc/fdo/keys/manufacturer_key.der",
		OwnerCertPath:          "/etc/fdo/keys/owner_cert.pem",
		DeviceCertCAPrivateKey: "/etc/fdo/keys/device_ca_key.der",
		DeviceCertCAChain:      "/etc/fdo/keys/device_ca_cert.pem",
	}
	return nil
}

func (c *ManufacturingServerConfig) setRendezvousValues(r []fdov1alpha1.RendezvousServer) error {
	rendezvousInfo := make([]RendezvousInfo, len(r))
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

func (c *ManufacturingServerConfig) setProtocolsValues(p *fdov1alpha1.Protocols) error {
	if !p.PlainDI && p.DIUN == nil {
		return fmt.Errorf("DIUN must be configured if plain DI is false")
	}

	if p.DIUN == nil {
		c.Protocols = &Protocols{
			PlainDI: p.PlainDI,
		}
		return nil
	}

	keyStorage := make([]string, len(p.DIUN.AllowedKeyStorageTypes))
	for i, t := range p.DIUN.AllowedKeyStorageTypes {
		keyStorage[i] = string(t)
	}

	c.Protocols = &Protocols{
		PlainDI: p.PlainDI,
		DIUN: &DIUN{
			KeyPath:                "/etc/fdo/keys/diun_key.der",
			CertPath:               "/etc/fdo/keys/diun_cert.pem",
			KeyType:                p.DIUN.KeyType,
			MFGStringType:          "SerialNumber",
			AllowedKeyStorageTypes: keyStorage,
		},
	}
	return nil
}

type RendezvousServerConfig struct {
	StorageDriver               *Driver `yaml:"storage_driver"`
	SessionStoreDriver          *Driver `yaml:"session_store_driver"`
	Bind                        string  `yaml:"bind"`
	TrustedManufacturerKeysPath string  `yaml:"trusted_manufacturer_keys_path"`
}

func (c *RendezvousServerConfig) setValues(s *fdov1alpha1.FDORendezvousServer) error {
	c.StorageDriver = NewDriver("/etc/fdo/rendezvous_registered/")
	c.SessionStoreDriver = NewDriver("/etc/fdo/rendezvous_sessions/")
	c.TrustedManufacturerKeysPath = "/etc/fdo/keys/manufacturer_cert.pem"
	c.Bind = "0.0.0.0:8082"
	return nil
}
