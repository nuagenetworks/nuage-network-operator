package nuagecniconfig

import (
	"context"
	"fmt"
	"net"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	iputil "github.com/nuagenetworks/nuage-network-operator/pkg/util/ip"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	//DefaultServiceNetworkCIDR is the default service cidr used by kubernetes as of v1.15
	DefaultServiceNetworkCIDR string = "10.0.0.0/24"
)

//GetK8SClusterNetworkInfo fetches service and pod cidr from api server
func (r *ReconcileNuageCNIConfig) GetK8SClusterNetworkInfo() (*operv1.ClusterNetworkConfigDefinition, error) {

	//if k8s, cluster network and cluster network subnet length
	// are read from crd directly and should have been populated by now
	c := &operv1.ClusterNetworkConfigDefinition{
		ClusterNetworkCIDR:         r.clusterNetworkCIDR,
		ClusterNetworkSubnetLength: r.clusterNetworkSubnetLength,
	}

	podList := &corev1.PodList{}
	lo := &client.ListOptions{Namespace: "kube-system"}
	lo.SetLabelSelector("component==kube-apiserver")

	err := r.client.List(context.TODO(), lo, podList)
	if err != nil {
		log.Errorf("fetching pod list failed")
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("api server pod could not be listed")
	}

	command := podList.Items[0].Spec.Containers[0].Command
	for _, arg := range command {
		if strings.Contains(arg, "service-cluster-ip-range") {
			kvs := strings.Split(arg, "=")
			if len(kvs) < 2 {
				c.ServiceNetworkCIDR = DefaultServiceNetworkCIDR
				break
			}
			c.ServiceNetworkCIDR = strings.Trim(kvs[1], `'"`)
		}
	}

	if len(c.ServiceNetworkCIDR) == 0 {
		c.ServiceNetworkCIDR = DefaultServiceNetworkCIDR
	}

	return c, nil
}

// GetClusterNetworkInfo fetches the cluster network configuration from API server
func (r *ReconcileNuageCNIConfig) GetClusterNetworkInfo(request reconcile.Request) (*operv1.ClusterNetworkConfigDefinition, error) {
	clusterConfig := &configv1.Network{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "network"}, clusterConfig)
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
		log.Errorf("Failed to validate Network.Spec %v", err)
		return nil, err
	}

	networkInfo := &operv1.ClusterNetworkConfigDefinition{
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
		return errors.Errorf("spec.serviceNetwork must have only one entry")
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

	if len(clusterConfig.ClusterNetwork) != 1 {
		return errors.Errorf("spec.clusterNetwork must have only one entry")
	}
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

	if clusterConfig.NetworkType != names.NuageSDN {
		return errors.Errorf("spec.networkType \"%s\"is not supported", clusterConfig.NetworkType)
	}

	return nil
}
