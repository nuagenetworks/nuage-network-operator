package monitor

import (
	"fmt"
	"net"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("vrs_config")

//Parse validates the Monitor config definition and fill in default values
func Parse(config *operv1.MonitorConfigDefinition) error {
	if err := validate(config); err != nil {
		return fmt.Errorf("validating monitor config failed %v", err)
	}

	fillDefaults(config)
	return nil
}

func validate(config *operv1.MonitorConfigDefinition) error {
	if ip := net.ParseIP(config.VSDAddress); ip == nil {
		return fmt.Errorf("invalid vsd ip address")
	}
	if config.VSDPort <= 0 {
		return fmt.Errorf("invalid vsd port address")
	}

	if err := validateMetadata(config.VSDMetadata); err != nil {
		return fmt.Errorf("vsd metadata validation failed: %v", err)
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
}
