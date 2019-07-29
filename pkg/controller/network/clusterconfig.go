package network

import (
	"context"
	"net"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	iputil "github.com/nuagenetworks/nuage-network-operator/pkg/util/ip"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var configlog = logf.Log.WithName("cluster_config")

// GetClusterNetworkInfo fetches the cluster network configuration from API server
func (r *ReconcileNetwork) GetClusterNetworkInfo(request reconcile.Request) (*operatorv1.ClusterNetworkConfigDefinition, error) {
	clusterConfig := &configv1.Network{}
	err := r.client.Get(context.TODO(), request.NamespacedName, clusterConfig)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil, nil
		}
		// Error reading the object - requeue the request.
		return nil, err
	}

	// Validate the cluster config
	if err = ValidateClusterConfig(clusterConfig.Spec); err != nil {
		configlog.Error(err, "Failed to validate Network.Spec")
		return nil, err
	}

	networkInfo := &operatorv1.ClusterNetworkConfigDefinition{
		ClusterNetworkCIDR:         clusterConfig.Spec.ClusterNetwork[0].CIDR,
		ServiceNetworkCIDR:         clusterConfig.Spec.ServiceNetwork[0],
		ClusterNetworkSubnetLength: clusterConfig.Spec.ClusterNetwork[0].HostPrefix,
	}
	return networkInfo, nil
}

// ValidateClusterConfig ensures the cluster config is valid.
func ValidateClusterConfig(clusterConfig configv1.NetworkSpec) error {
	// Check all networks for overlaps
	pool := iputil.IPPool{}

	if len(clusterConfig.ServiceNetwork) != 1 {
		// Right now we only support a single service network
		return errors.Errorf("spec.serviceNetwork must be exactly one network")
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

	if len(clusterConfig.ClusterNetwork) != 1 {
		return errors.Errorf("spec.clusterNetwork must have 1 entry")
	}

	if clusterConfig.NetworkType != names.NuageSDN {
		return errors.Errorf("spec.networkType is not supported")
	}

	return nil
}
