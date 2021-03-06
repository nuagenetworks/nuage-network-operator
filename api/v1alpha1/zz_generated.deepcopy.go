// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

// +build !ignore_autogenerated

/*


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

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CNIConfigDefinition) DeepCopyInto(out *CNIConfigDefinition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CNIConfigDefinition.
func (in *CNIConfigDefinition) DeepCopy() *CNIConfigDefinition {
	if in == nil {
		return nil
	}
	out := new(CNIConfigDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertGenConfig) DeepCopyInto(out *CertGenConfig) {
	*out = *in
	if in.ECDSACurve != nil {
		in, out := &in.ECDSACurve, &out.ECDSACurve
		*out = new(string)
		**out = **in
	}
	if in.ValidFrom != nil {
		in, out := &in.ValidFrom, &out.ValidFrom
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertGenConfig.
func (in *CertGenConfig) DeepCopy() *CertGenConfig {
	if in == nil {
		return nil
	}
	out := new(CertGenConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterNetworkConfigDefinition) DeepCopyInto(out *ClusterNetworkConfigDefinition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterNetworkConfigDefinition.
func (in *ClusterNetworkConfigDefinition) DeepCopy() *ClusterNetworkConfigDefinition {
	if in == nil {
		return nil
	}
	out := new(ClusterNetworkConfigDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Flags) DeepCopyInto(out *Flags) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Flags.
func (in *Flags) DeepCopy() *Flags {
	if in == nil {
		return nil
	}
	out := new(Flags)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InfraPodConfigDefenition) DeepCopyInto(out *InfraPodConfigDefenition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InfraPodConfigDefenition.
func (in *InfraPodConfigDefenition) DeepCopy() *InfraPodConfigDefenition {
	if in == nil {
		return nil
	}
	out := new(InfraPodConfigDefenition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metadata) DeepCopyInto(out *Metadata) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metadata.
func (in *Metadata) DeepCopy() *Metadata {
	if in == nil {
		return nil
	}
	out := new(Metadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitorConfigDefinition) DeepCopyInto(out *MonitorConfigDefinition) {
	*out = *in
	out.VSDMetadata = in.VSDMetadata
	out.VSDFlags = in.VSDFlags
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitorConfigDefinition.
func (in *MonitorConfigDefinition) DeepCopy() *MonitorConfigDefinition {
	if in == nil {
		return nil
	}
	out := new(MonitorConfigDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuageCNIConfig) DeepCopyInto(out *NuageCNIConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuageCNIConfig.
func (in *NuageCNIConfig) DeepCopy() *NuageCNIConfig {
	if in == nil {
		return nil
	}
	out := new(NuageCNIConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NuageCNIConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuageCNIConfigList) DeepCopyInto(out *NuageCNIConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NuageCNIConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuageCNIConfigList.
func (in *NuageCNIConfigList) DeepCopy() *NuageCNIConfigList {
	if in == nil {
		return nil
	}
	out := new(NuageCNIConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NuageCNIConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuageCNIConfigSpec) DeepCopyInto(out *NuageCNIConfigSpec) {
	*out = *in
	in.VRSConfig.DeepCopyInto(&out.VRSConfig)
	out.CNIConfig = in.CNIConfig
	out.MonitorConfig = in.MonitorConfig
	out.ReleaseConfig = in.ReleaseConfig
	out.PodNetworkConfig = in.PodNetworkConfig
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuageCNIConfigSpec.
func (in *NuageCNIConfigSpec) DeepCopy() *NuageCNIConfigSpec {
	if in == nil {
		return nil
	}
	out := new(NuageCNIConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NuageCNIConfigStatus) DeepCopyInto(out *NuageCNIConfigStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NuageCNIConfigStatus.
func (in *NuageCNIConfigStatus) DeepCopy() *NuageCNIConfigStatus {
	if in == nil {
		return nil
	}
	out := new(NuageCNIConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodNetworkConfigDefinition) DeepCopyInto(out *PodNetworkConfigDefinition) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodNetworkConfigDefinition.
func (in *PodNetworkConfigDefinition) DeepCopy() *PodNetworkConfigDefinition {
	if in == nil {
		return nil
	}
	out := new(PodNetworkConfigDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegistryConfig) DeepCopyInto(out *RegistryConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegistryConfig.
func (in *RegistryConfig) DeepCopy() *RegistryConfig {
	if in == nil {
		return nil
	}
	out := new(RegistryConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReleaseConfigDefinition) DeepCopyInto(out *ReleaseConfigDefinition) {
	*out = *in
	out.Registry = in.Registry
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReleaseConfigDefinition.
func (in *ReleaseConfigDefinition) DeepCopy() *ReleaseConfigDefinition {
	if in == nil {
		return nil
	}
	out := new(ReleaseConfigDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RenderConfig) DeepCopyInto(out *RenderConfig) {
	*out = *in
	in.NuageCNIConfigSpec.DeepCopyInto(&out.NuageCNIConfigSpec)
	if in.Certificates != nil {
		in, out := &in.Certificates, &out.Certificates
		*out = new(TLSCertificates)
		(*in).DeepCopyInto(*out)
	}
	if in.ClusterNetworkConfig != nil {
		in, out := &in.ClusterNetworkConfig, &out.ClusterNetworkConfig
		*out = new(ClusterNetworkConfigDefinition)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RenderConfig.
func (in *RenderConfig) DeepCopy() *RenderConfig {
	if in == nil {
		return nil
	}
	out := new(RenderConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSCertificates) DeepCopyInto(out *TLSCertificates) {
	*out = *in
	if in.CA != nil {
		in, out := &in.CA, &out.CA
		*out = new(string)
		**out = **in
	}
	if in.Certificate != nil {
		in, out := &in.Certificate, &out.Certificate
		*out = new(string)
		**out = **in
	}
	if in.PrivateKey != nil {
		in, out := &in.PrivateKey, &out.PrivateKey
		*out = new(string)
		**out = **in
	}
	if in.CertificateDir != nil {
		in, out := &in.CertificateDir, &out.CertificateDir
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSCertificates.
func (in *TLSCertificates) DeepCopy() *TLSCertificates {
	if in == nil {
		return nil
	}
	out := new(TLSCertificates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VRSConfigDefinition) DeepCopyInto(out *VRSConfigDefinition) {
	*out = *in
	if in.Controllers != nil {
		in, out := &in.Controllers, &out.Controllers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VRSConfigDefinition.
func (in *VRSConfigDefinition) DeepCopy() *VRSConfigDefinition {
	if in == nil {
		return nil
	}
	out := new(VRSConfigDefinition)
	in.DeepCopyInto(out)
	return out
}
