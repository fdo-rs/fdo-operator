//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Address) DeepCopyInto(out *Address) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Address.
func (in *Address) DeepCopy() *Address {
	if in == nil {
		return nil
	}
	out := new(Address)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOManufacturingServer) DeepCopyInto(out *FDOManufacturingServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOManufacturingServer.
func (in *FDOManufacturingServer) DeepCopy() *FDOManufacturingServer {
	if in == nil {
		return nil
	}
	out := new(FDOManufacturingServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FDOManufacturingServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOManufacturingServerList) DeepCopyInto(out *FDOManufacturingServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FDOManufacturingServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOManufacturingServerList.
func (in *FDOManufacturingServerList) DeepCopy() *FDOManufacturingServerList {
	if in == nil {
		return nil
	}
	out := new(FDOManufacturingServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FDOManufacturingServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOManufacturingServerSpec) DeepCopyInto(out *FDOManufacturingServerSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOManufacturingServerSpec.
func (in *FDOManufacturingServerSpec) DeepCopy() *FDOManufacturingServerSpec {
	if in == nil {
		return nil
	}
	out := new(FDOManufacturingServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOManufacturingServerStatus) DeepCopyInto(out *FDOManufacturingServerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOManufacturingServerStatus.
func (in *FDOManufacturingServerStatus) DeepCopy() *FDOManufacturingServerStatus {
	if in == nil {
		return nil
	}
	out := new(FDOManufacturingServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOOnboardingServer) DeepCopyInto(out *FDOOnboardingServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOOnboardingServer.
func (in *FDOOnboardingServer) DeepCopy() *FDOOnboardingServer {
	if in == nil {
		return nil
	}
	out := new(FDOOnboardingServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FDOOnboardingServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOOnboardingServerList) DeepCopyInto(out *FDOOnboardingServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FDOOnboardingServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOOnboardingServerList.
func (in *FDOOnboardingServerList) DeepCopy() *FDOOnboardingServerList {
	if in == nil {
		return nil
	}
	out := new(FDOOnboardingServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FDOOnboardingServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOOnboardingServerSpec) DeepCopyInto(out *FDOOnboardingServerSpec) {
	*out = *in
	if in.OwnerAddresses != nil {
		in, out := &in.OwnerAddresses, &out.OwnerAddresses
		*out = make([]OwnerAddress, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOOnboardingServerSpec.
func (in *FDOOnboardingServerSpec) DeepCopy() *FDOOnboardingServerSpec {
	if in == nil {
		return nil
	}
	out := new(FDOOnboardingServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDOOnboardingServerStatus) DeepCopyInto(out *FDOOnboardingServerStatus) {
	*out = *in
	if in.Pods != nil {
		in, out := &in.Pods, &out.Pods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDOOnboardingServerStatus.
func (in *FDOOnboardingServerStatus) DeepCopy() *FDOOnboardingServerStatus {
	if in == nil {
		return nil
	}
	out := new(FDOOnboardingServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDORendezvousServer) DeepCopyInto(out *FDORendezvousServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDORendezvousServer.
func (in *FDORendezvousServer) DeepCopy() *FDORendezvousServer {
	if in == nil {
		return nil
	}
	out := new(FDORendezvousServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FDORendezvousServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDORendezvousServerList) DeepCopyInto(out *FDORendezvousServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FDORendezvousServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDORendezvousServerList.
func (in *FDORendezvousServerList) DeepCopy() *FDORendezvousServerList {
	if in == nil {
		return nil
	}
	out := new(FDORendezvousServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FDORendezvousServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDORendezvousServerSpec) DeepCopyInto(out *FDORendezvousServerSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDORendezvousServerSpec.
func (in *FDORendezvousServerSpec) DeepCopy() *FDORendezvousServerSpec {
	if in == nil {
		return nil
	}
	out := new(FDORendezvousServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FDORendezvousServerStatus) DeepCopyInto(out *FDORendezvousServerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FDORendezvousServerStatus.
func (in *FDORendezvousServerStatus) DeepCopy() *FDORendezvousServerStatus {
	if in == nil {
		return nil
	}
	out := new(FDORendezvousServerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OwnerAddress) DeepCopyInto(out *OwnerAddress) {
	*out = *in
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]Address, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OwnerAddress.
func (in *OwnerAddress) DeepCopy() *OwnerAddress {
	if in == nil {
		return nil
	}
	out := new(OwnerAddress)
	in.DeepCopyInto(out)
	return out
}
