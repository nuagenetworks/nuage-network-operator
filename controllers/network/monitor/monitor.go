// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package monitor

import (
	"fmt"

	operv1 "github.com/nuagenetworks/nuage-network-operator/api/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/controllers/names"
)

var (
	//DefaultResourceName is the default resource name for all resources involved in RBAC
	DefaultResourceName = "nuage-monitor"
	//DefaultRestServerAddress is the default rest server address
	DefaultRestServerAddress = "0.0.0.0"
	//DefaultRestServerPort is the default rest server port
	DefaultRestServerPort = 9443
)

//Parse validates the Monitor config definition and fill in default values
func Parse(config *operv1.MonitorConfigDefinition) error {
	if err := validate(config); err != nil {
		return fmt.Errorf("validating monitor config failed %v", err)
	}

	fillDefaults(config)
	return nil
}

func validate(config *operv1.MonitorConfigDefinition) error {
	if len(config.VSDAddress) == 0 {
		return fmt.Errorf("invalid vsd ip address")
	}

	if config.VSDPort <= 0 {
		return fmt.Errorf("invalid vsd port address")
	}

	if err := validateMetadata(config.VSDMetadata); err != nil {
		return fmt.Errorf("vsd metadata validation failed: %v", err)
	}

	if config.RestServerPort < 0 {
		return fmt.Errorf("invalid rest server port")
	}
	return nil
}

func validateMetadata(m operv1.Metadata) error {
	if len(m.Enterprise) == 0 {
		return fmt.Errorf("enterprise name cannot be empty")
	}
	if len(m.Domain) == 0 {
		return fmt.Errorf("domain name cannot be empty")
	}
	if len(m.User) == 0 {
		return fmt.Errorf("user name cannot be empty")
	}
	if len(m.UserCert) == 0 {
		return fmt.Errorf("user certificate cannot be empty")
	}
	if len(m.UserKey) == 0 {
		return fmt.Errorf("user key cannot be empty")
	}
	return nil
}

func fillDefaults(config *operv1.MonitorConfigDefinition) {
	//config.VSDFlags are all boolean. They default to false
	//which we want
	if len(config.RestServerAddress) == 0 {
		config.RestServerAddress = DefaultRestServerAddress
	}

	if config.RestServerPort == 0 {
		config.RestServerPort = DefaultRestServerPort
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

	if len(config.MasterNodeSelector) == 0 {
		config.MasterNodeSelector = names.MasterNodeSelector
	}
}
