package cni

import (
	"fmt"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("cni_config")

const (
	//VRSSocketFile socker file for connection to VRS
	VRSSocketFile = "/var/run/openvswitch/db.sock"
	//DefaultVRSBridge default vrs bridge to add vports
	DefaultVRSBridge = "alubr0"
	//CNIVersion version that the CNI plugin uses
	CNIVersion = "0.2.0"
	//LogLevel verbosity level for log messages
	LogLevel = "info"
	//MTU setting the interface MTU for packets
	MTU = 1450
	//NuageSiteID to be used for EVDF personalities
	NuageSiteID = -1
	//LogFileSize maximum file size after which rotation happens
	LogFileSize = 1
	//MonitorInterval monitor polling interval
	MonitorInterval = 60
	//PortResolveTimer timeout after which port resolution is declared failure
	PortResolveTimer = 60
	//VRSConnCheckTimer timeout after which vrs connection is re initiated
	VRSConnCheckTimer = 180
	//StaleEntryTimeout removes the stale entries from OVSDB after this time
	StaleEntryTimeout = 600
	//DefaultResourceName is the name of the resources like sa, role and role binding
	DefaultResourceName = "nuage-cni"
)

//Parse validates the CNI config definition and fill in default values
func Parse(config *operv1.CNIConfigDefinition) error {
	if err := validate(config); err != nil {
		log.Error(err, "validating vrs config failed")
		return err
	}

	fillDefaults(config)
	return nil
}

func validate(config *operv1.CNIConfigDefinition) error {
	if config.MTU > 1450 {
		return fmt.Errorf("mtu exceeds 1450")
	}
	if config.NuageSiteID > 0 {
		return fmt.Errorf("non negative values of site id is not supported")
	}
	if len(config.LoadBalancerURL) == 0 {
		return fmt.Errorf("load balancer url cannot be empty")
	}
	return nil
}

func fillDefaults(config *operv1.CNIConfigDefinition) {
	if len(config.VRSEndpoint) == 0 {
		config.VRSEndpoint = VRSSocketFile
	}
	if len(config.VRSBridge) == 0 {
		config.VRSBridge = DefaultVRSBridge
	}
	if len(config.CNIVersion) == 0 {
		config.CNIVersion = CNIVersion
	}
	if len(config.LogLevel) == 0 {
		config.LogLevel = LogLevel
	}
	if config.MTU == 0 {
		config.MTU = MTU
	}
	if config.NuageSiteID == 0 {
		config.NuageSiteID = -1
	}
	if config.LogFileSize == 0 {
		config.LogFileSize = 1
	}
	if config.MonitorInterval == 0 {
		config.MonitorInterval = 60
	}
	if config.PortResolveTimer == 0 {
		config.PortResolveTimer = 60
	}
	if config.VRSConnectionCheckTimer == 0 {
		config.VRSConnectionCheckTimer = 180
	}
	if config.StaleEntryTimeout == 0 {
		config.StaleEntryTimeout = 600
	}
	if len(config.ServiceAccountName) == 0 {
		config.ServiceAccountName = DefaultResourceName
	}

	if len(config.ClusterRoleName) == 0 {
		config.ClusterRoleName = DefaultResourceName
	}

	if len(config.ClusterRoleBindingName) == 0 {
		config.ClusterRoleBindingName = DefaultResourceName
	}

}
