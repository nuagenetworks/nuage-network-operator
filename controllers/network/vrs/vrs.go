// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package vrs

import (
	"fmt"
	"net"

	operv1 "github.com/nuagenetworks/nuage-network-operator/api/v1alpha1"
)

const (
	//VRSPlatform defines the VRS platform
	VRSPlatform = "kvm, k8s"
)

//Parse validates the VRS config definition and fill in default values
func Parse(config *operv1.VRSConfigDefinition) error {
	if err := validate(config); err != nil {
		return fmt.Errorf("validating vrs config failed %v", err)
	}

	fillDefaults(config)
	return nil
}

func validate(config *operv1.VRSConfigDefinition) error {
	if len(config.Controllers) == 0 {
		return fmt.Errorf("atleast one controller is expected")
	}
	for _, controllerIP := range config.Controllers {
		ip := net.ParseIP(controllerIP)
		if ip == nil {
			return fmt.Errorf("controller ip is not valid")
		}
	}

	if len(config.UnderlayUplink) == 0 {
		return fmt.Errorf("underlay uplink cannot be empty")
	}
	return nil
}

func fillDefaults(config *operv1.VRSConfigDefinition) {
	if len(config.Platform) == 0 {
		config.Platform = VRSPlatform
	}
}
