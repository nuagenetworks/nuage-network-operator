package vrs

import (
	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("vrs_config")

//Parse validates the VRS config definition and fill in default values
func Parse(config *operv1.VRSConfigDefinition) error {
	if err := validate(config); err != nil {
		log.Error(err, "validating vrs config failed")
		return err
	}

	fillDefaults(config)
	return nil
}

func validate(config *operv1.VRSConfigDefinition) error {
	return nil
}

func fillDefaults(config *operv1.VRSConfigDefinition) {
}
