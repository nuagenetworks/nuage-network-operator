## Nuage CNI Operator

Nuage CNI operator manages Nuage Monitor, VRS and CNI daemonsets in a Kubernetes/OpenShift cluster. This operator reconciles Nuage CNI config custom resource and creates/updates daemonsets based on the cluster state before reconcile.

### Building the operator

Nuage CNI Operator images can be built using [operator-sdk](https://github.com/operator-framework/operator-sdk) version v0.17.0

    operator-sdk build <image name>:<tag>

### Kubernetes

Please refer to offical Nuage documentation with detailed information to create a Kubernetes cluster with Nuage SDN.

Following steps only provide an overview to create a Kubernetes cluster with Nuage SDN as networking backend. 

1. Create initial kubernetes cluster using [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/). Nodes would be in NotReady state as the network components are not yet created.
2. On VSD, create an enterprise and add an admin user to the enterprise. Please refer to VSP documentation for this
3. Update operator image in the [deployment](./deploy/005-operator.yaml)
4. Deploy Operator and related artifacts using `kubectl apply -f deploy/`
5. Populate NuageCNIConfig custom resource. A sample custom resource file can be found [here](./deploy/crds/operator_v1alpha1_nuagecniconfig_cr.yaml)
6. Nuage Monitor, CNI and VRS components are created in `nuage-network-operator` namespaces as daemonsets
