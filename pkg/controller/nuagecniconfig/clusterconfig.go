// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package nuagecniconfig

import (
	"context"
	"net"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/cni"
	iputil "github.com/nuagenetworks/nuage-network-operator/pkg/util/ip"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	//DefaultServiceNetworkCIDR is the default service cidr used by kubernetes as of v1.15
	DefaultServiceNetworkCIDR string = "10.0.0.0/24"
)

// GetClusterNetworkInfo fetches the cluster network configuration from API server
func (r *ReconcileNuageCNIConfig) GetClusterNetworkInfo() (*operv1.ClusterNetworkConfigDefinition, error) {
	if r.orchestrator == OrchestratorKubernetes {
		return r.GetK8SClusterNetworkInfo()
	}

	return r.GetOSEClusterNetworkInfo()
}

//GetK8SClusterNetworkInfo fetches service cidr from k8s api server
func (r *ReconcileNuageCNIConfig) GetK8SClusterNetworkInfo() (*operv1.ClusterNetworkConfigDefinition, error) {

	//if k8s, cluster network and cluster network subnet length
	// are read from crd directly and should have been populated by now
	c := &operv1.ClusterNetworkConfigDefinition{
		ClusterNetworkCIDR:         r.clusterNetworkCIDR,
		ClusterNetworkSubnetLength: r.clusterNetworkSubnetLength,
		ServiceNetworkCIDR:         r.ClusterServiceNetworkCIDR,
	}

	if len(r.ClusterServiceNetworkCIDR) == 0 {
		c.ServiceNetworkCIDR = DefaultServiceNetworkCIDR
	}

	return c, nil
}

// GetOSEClusterNetworkInfo fetches network config from api server
func (r *ReconcileNuageCNIConfig) GetOSEClusterNetworkInfo() (*operv1.ClusterNetworkConfigDefinition, error) {
	clusterConfig := &configv1.Network{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "network"}, clusterConfig)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, err
		}
		return nil, err
	}

	// Validate the cluster config
	if err = ValidateOSEClusterConfig(clusterConfig.Spec); err != nil {
		log.Errorf("Failed to validate Network.Spec %v", err)
		return nil, nil
	}

	networkInfo := &operv1.ClusterNetworkConfigDefinition{
		ClusterNetworkCIDR:         clusterConfig.Spec.ClusterNetwork[0].CIDR,
		ServiceNetworkCIDR:         clusterConfig.Spec.ServiceNetwork[0],
		ClusterNetworkSubnetLength: clusterConfig.Spec.ClusterNetwork[0].HostPrefix,
	}
	return networkInfo, nil
}

// ValidateOSEClusterConfig ensures the cluster config is valid.
func ValidateOSEClusterConfig(clusterConfig configv1.NetworkSpec) error {
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
			return errors.Errorf("subnet length %d is larger than its cidr %s",
				cnet.HostPrefix, cnet.CIDR)
		}
		if cnet.HostPrefix > 30 {
			return errors.Errorf("subnet length %d is too small, must be a /30 or larger",
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

// ValidateK8SClusterConfig validates the cluster config for k8s
func ValidateK8SClusterConfig(c *operv1.ClusterNetworkConfigDefinition) error {
	// Check all networks for overlaps
	pool := iputil.IPPool{}

	_, cidr, err := net.ParseCIDR(c.ServiceNetworkCIDR)
	if err != nil {
		return errors.Errorf("invalid service network cidr found %v", c.ServiceNetworkCIDR)
	}

	if err := pool.Add(*cidr); err != nil {
		return err
	}

	_, cidr, err = net.ParseCIDR(c.ClusterNetworkCIDR)
	if err != nil {
		return errors.Errorf("invalid pod network cidr found %v", c.ClusterNetworkCIDR)
	}

	if err := pool.Add(*cidr); err != nil {
		return err
	}

	size, _ := cidr.Mask.Size()
	// The comparison is inverted; smaller number is larger block
	if c.ClusterNetworkSubnetLength < uint32(size) {
		return errors.Errorf("subnet length %d is larger than its cidr %s",
			c.ClusterNetworkSubnetLength, c.ClusterNetworkCIDR)
	}
	if c.ClusterNetworkSubnetLength > 30 {
		return errors.Errorf("subnet length %d is too small, must be a /30 or larger",
			c.ClusterNetworkSubnetLength)
	}
	return nil
}

//UpdateClusterNetworkStatus updates config.openshift.io/v1 status object
func (r *ReconcileNuageCNIConfig) UpdateClusterNetworkStatus(c *operv1.ClusterNetworkConfigDefinition) error {
	if r.orchestrator == OrchestratorKubernetes {
		return nil
	}

	clusterConfig := &configv1.Network{
		TypeMeta:   metav1.TypeMeta{APIVersion: configv1.GroupVersion.String(), Kind: "Network"},
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
		Status: configv1.NetworkStatus{
			ClusterNetwork: []configv1.ClusterNetworkEntry{
				{
					CIDR:       c.ClusterNetworkCIDR,
					HostPrefix: c.ClusterNetworkSubnetLength,
				},
			},
			ServiceNetwork:    []string{c.ServiceNetworkCIDR},
			NetworkType:       names.NuageSDN,
			ClusterNetworkMTU: cni.MTU,
		},
	}

	return r.ApplyObject(types.NamespacedName{Name: clusterConfig.GetName()}, clusterConfig)
}
