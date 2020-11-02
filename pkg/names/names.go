// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package names

const (
	// NuageSDN Network Type
	NuageSDN = "NuageSDN"
	// Namespace is the namespace in which nuage network operator is deployed
	Namespace = "nuage-network-operator"
	// NuageReleaseConfig is name of the config map used to store release config
	NuageReleaseConfig = "nuage-release-config"
	// NuageCertConfig is name of the config map used to store release config
	NuageCertConfig = "nuage-cert-config"
	// ServiceAccountName is the name of the service account used for cni
	ServiceAccountName = "nuage-network-operator"
	// MasterNodeSelector label to be used for selecting master nodes
	MasterNodeSelector = "nuage.io/monitor-pod"
	NuageMonitorConfig = "nuage-monitor-config-data"
	NuageMonitor       = "nuage-monitor"
)
