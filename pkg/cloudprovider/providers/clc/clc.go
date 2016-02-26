/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clc

import (
	"errors"
	"net"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/types"
)

const (
	ProviderName = "clc"
)

// CLCCloud is an implementation of Interface, LoadBalancer and Instances for CenturyLinkCloud.
type CLCCloud struct {
}

// LoadBalancer returns a balancer interface. Also returns true if the interface is supported, false otherwise.
func (clc *CLCCloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return nil, false
}

// Instances returns an instances interface. Also returns true if the interface is supported, false otherwise.
func (clc *CLCCloud) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

// Zones returns a zones interface. Also returns true if the interface is supported, false otherwise.
func (clc *CLCCloud) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

// Clusters returns a clusters interface.  Also returns true if the interface is supported, false otherwise.
func (clc *CLCCloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes returns a routes interface along with whether the interface is supported.
func (clc *CLCCloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// ProviderName returns the cloud provider ID.
func (clc *CLCCloud) ProviderName() string {
	return ProviderName
}

// ScrubDNS provides an opportunity for cloud-provider-specific code to process DNS settings for pods.
func (clc *CLCCloud) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nil, nil
}

// ListClusters lists the names of the available clusters.
func (clc *CLCCloud) ListClusters() ([]string, error) {
	return nil, errors.New("unsupported method")
}

// Master gets back the address (either DNS name or IP address) of the master node for the cluster.
func (clc *CLCCloud) Master(clusterName string) (string, error) {
	return "", errors.New("unsupported method")
}

// NodeAddresses returns the addresses of the specified instance.
func (clc *CLCCloud) NodeAddresses(name string) ([]api.NodeAddress, error) {
	return nil, errors.New("unsupported method")
}

// ExternalID returns the cloud provider ID of the specified instance (deprecated).
func (clc *CLCCloud) ExternalID(name string) (string, error) {
	return "", errors.New("unsupported method")
}

// InstanceID returns the cloud provider ID of the specified instance.
func (clc *CLCCloud) InstanceID(name string) (string, error) {
	return "", errors.New("unsupported method")
}

// InstanceType returns the type of the specified instance.
func (clc *CLCCloud) InstanceType(name string) (string, error) {
	return "", errors.New("unsupported method")
}

// List lists instances that match 'filter' which is a regular expression which must match the entire instance name (fqdn)
func (clc *CLCCloud) List(filter string) ([]string, error) {
	return nil, errors.New("unsupported method")
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
func (clc *CLCCloud) AddSSHKeyToAllInstances(user string, keyData []byte) error {
	return errors.New("unsupported method")
}

// CurrentNodeName returns the name of the node we are currently running on
func (clc *CLCCloud) CurrentNodeName(hostname string) (string, error) {
	return "", errors.New("unsupported method")
}

// GetLoadBalancer returns whether the specified load balancer exists, and
// if so, what its status is.
func (clc *CLCCloud) GetLoadBalancer(name, region string) (status *api.LoadBalancerStatus, exists bool, err error) {
	return nil, false, errors.New("unsupported method")
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
func (clc *CLCCloud) EnsureLoadBalancer(name, region string, loadBalancerIP net.IP, ports []*api.ServicePort, hosts []string, serviceName types.NamespacedName, affinityType api.ServiceAffinity, annotations cloudprovider.ServiceAnnotation) (*api.LoadBalancerStatus, error) {
	return nil, errors.New("unsupported method")
}

// UpdateLoadBalancer updates hosts under the specified load balancer.
func (clc *CLCCloud) UpdateLoadBalancer(name, region string, hosts []string) error {
	return errors.New("unsupported method")
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
func (clc *CLCCloud) EnsureLoadBalancerDeleted(name, region string) error {
	return errors.New("unsupported method")
}
