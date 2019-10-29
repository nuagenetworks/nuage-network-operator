## Nuage CNI Operator

Nuage CNI operator manages Nuage Monitor, VRS and CNI daemonsets in a Kubernetes/OpenShift cluster. This operator reconciles Nuage CNI config custom resource and creates/updates daemonsets based on the cluster state before reconcile.

### Kuberentes

Following steps can be used to create a Kubernetes cluster with Nuage SDN as networking backend.

1. Create initial kubernetes cluster using [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/). Nodes would be in NotReady state as the network components are not yet created.
2. On VSD, create an enterprise and add an admin user to the enterprise. Please refer to VSP documentation for this
3. Update operator image in the [deployment](./deploy/005-operator.yaml)
4. Deploy Operator and related artifacts using `kubectl apply -f deploy/`
5. Populate NuageCNIConfig custom resource. A sample custom resource file can be found [here](./deploy/crds/operator_v1alpha1_nuagecniconfig_cr.yaml)
6. Nuage Monitor, CNI and VRS components are created in `nuage-network-operator` namespaces as daemonsets

### OpenShift

Following steps can be used to create OpenShift 4.x cluster with Nuage SDN as networking backend. These steps are taken from OpenShift documentation and modified to suit to Nuage SDN install. Please refer to [link1](https://docs.openshift.com/container-platform/4.1/installing/installing_bare_metal/installing-bare-metal.html) and [link2](https://redhat-connect.gitbook.io/certified-operator-guide/appendix/using-third-party-network-operators-with-openshift) for more information

1. Create a work directory

    mkdir mycluster

2. Create install config

    openshift-install create install-config --dir=mycluster

3. Generate the manifests

    openshift-install create manifests --dir=mycluster

4. Copy the operator [manifests](./deploy) to the installer

5. Create the cluster using the remaining steps in this [document](https://docs.openshift.com/container-platform/4.1/installing/installing_bare_metal/installing-bare-metal.html)
