package network

import (
	"net"

	iputil "github.com/nuagenetworks/nuage-network-operator/pkg/util/ip"
	configv1 "github.com/openshift/api/config/v1"

	"github.com/pkg/errors"
)

// ValidateClusterConfig ensures the cluster config is valid.
func ValidateClusterConfig(clusterConfig configv1.NetworkSpec) error {
	// Check all networks for overlaps
	pool := iputil.IPPool{}

	if len(clusterConfig.ServiceNetwork) == 0 {
		// Right now we only support a single service network
		return errors.Errorf("spec.serviceNetwork must have at least 1 entry")
	}
	for _, snet := range clusterConfig.ServiceNetwork {
		_, cidr, err := net.ParseCIDR(snet)
		if err != nil {
			return errors.Wrapf(err, "could not parse spec.serviceNetwork %s", snet)
		}
		if err := pool.Add(*cidr); err != nil {
			return err
		}
	}

	// validate clusternetwork
	// - has an entry
	// - it is a valid ip
	// - has a reasonable cidr
	// - they do not overlap and do not overlap with the service cidr
	for _, cnet := range clusterConfig.ClusterNetwork {
		_, cidr, err := net.ParseCIDR(cnet.CIDR)
		if err != nil {
			return errors.Errorf("could not parse spec.clusterNetwork %s", cnet.CIDR)
		}
		size, _ := cidr.Mask.Size()
		// The comparison is inverted; smaller number is larger block
		if cnet.HostPrefix < uint32(size) {
			return errors.Errorf("hostPrefix %d is larger than its cidr %s",
				cnet.HostPrefix, cnet.CIDR)
		}
		if cnet.HostPrefix > 30 {
			return errors.Errorf("hostPrefix %d is too small, must be a /30 or larger",
				cnet.HostPrefix)
		}
		if err := pool.Add(*cidr); err != nil {
			return err
		}
	}

	if len(clusterConfig.ClusterNetwork) < 1 {
		return errors.Errorf("spec.clusterNetwork must have at least 1 entry")
	}

	if clusterConfig.NetworkType == "" {
		return errors.Errorf("spec.networkType is required")
	}

	return nil
}
