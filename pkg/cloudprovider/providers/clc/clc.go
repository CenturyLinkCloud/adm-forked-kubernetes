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
	"fmt"
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/scalingdata/gcfg"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/types"
)

const (
	// ProviderName clc for CenturyLinkCloud
	ProviderName = "clc"
)

// CLCCloud is an implementation of Interface, LoadBalancer and Instances for CenturyLinkCloud.
type CLCCloud struct {
	clcClient CenturyLinkClient // Q: how is this constructed?  Who makes a CLCCloud instance?
	clcLB     *clcProviderLB    // cloudprovider's LoadBalancer interface is implemented here
}

func init() {
	glog.Info("registering CLC provider")

	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := readConfig(config)
		if err != nil {
			return nil, err
		}
		return newCLCCloud(cfg)
	})
}

// Config holds CenturyLinkCloud configuration parameters
type Config struct {
	Global struct {
		Username   string `gcfg:"username"`
		Password   string `gcfg:"password"`
		Alias      string `gcfg:"alias"`
		Token      string
		Datacenter string `gcfg:"datacenter"`
	}
	// LoadBalancer LoadBalancerOpts
}

func readConfig(config io.Reader) (Config, error) {
	if config == nil {
		err := fmt.Errorf("no CenturyLinkCloud provider config file given")
		return Config{}, err
	}

	var cfg Config
	err := gcfg.ReadInto(&cfg, config)
	return cfg, err
}

func newCLCCloud(cfg Config) (*CLCCloud, error) {
	newClient, error := ClientLogin(cfg.Global.Username, cfg.Global.Password)
	if error != nil {
		return &CLCCloud{}, error
		// nyi return successfully with an empty creds object.  Maybe it'll work later
	}

	return &CLCCloud{
		clcClient: newClient,
		clcLB:     makeProviderLB(newClient),
	}, nil
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

//////////////// Kubernetes LoadBalancer interface: Get, Ensure, Update, EnsureDeleted
func (clc *CLCCloud) GetLoadBalancer(name, region string) (status *api.LoadBalancerStatus, exists bool, err error) {
	return clc.clcLB.GetLoadBalancer(name, region)
}

func (clc *CLCCloud) EnsureLoadBalancer(name, region string, loadBalancerIP net.IP,
	ports []*api.ServicePort, hosts []string, serviceName types.NamespacedName,
	affinityType api.ServiceAffinity, annotations cloudprovider.ServiceAnnotation) (*api.LoadBalancerStatus, error) {

	return clc.clcLB.EnsureLoadBalancer(name, region, loadBalancerIP, ports, hosts, serviceName, affinityType, annotations)
}

func (clc *CLCCloud) UpdateLoadBalancer(name, region string, hosts []string) error {
	return clc.clcLB.UpdateLoadBalancer(name, region, hosts)
}

func (clc *CLCCloud) EnsureLoadBalancerDeleted(name, region string) error {
	return clc.clcLB.EnsureLoadBalancerDeleted(name, region)
}

// GetZone returns the Zone containing the current failure zone and locality region that the program is running in
func (clc *CLCCloud) GetZone() (cloudprovider.Zone, error) {
	return cloudprovider.Zone{}, errors.New("unsupported method")
}

// ListRoutes lists all managed routes that belong to the specified clusterName
func (clc *CLCCloud) ListRoutes(clusterName string) ([]*cloudprovider.Route, error) {
	return nil, errors.New("unsupported method")
}

// CreateRoute creates the described managed route
func (clc *CLCCloud) CreateRoute(clusterName string, nameHint string, route *cloudprovider.Route) error {
	return errors.New("unsupported method")
}

// DeleteRoute deletes the specified managed route
// Route should be as returned by ListRoutes
func (clc *CLCCloud) DeleteRoute(clusterName string, route *cloudprovider.Route) error {
	return errors.New("unsupported method")
}
