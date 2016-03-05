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
	clcClient CenturyLinkClient	// Q: how is this constructed?  Who makes a CLCCloud instance?
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
// [OPEN] send this code to a separate file, leaving wrapper functions here so that this file is strictly a glue layer

func findLoadBalancerInstance(clcClient CenturyLinkClient, name, region string)	(*LoadBalancerDetails, error)  {
	// name is the Kubernetes-assigned name.  EnsureLoadBalancer assigns that to our LoadBalancerDetails.Name

	summaries,err := clcClient.listAllLB()
	if err != nil {
		return nil, err
	}

	for _,lbSummary := range summaries {
		if ((lbSummary.DataCenter == region) && (lbSummary.Name == name)) {
			return clcClient.inspectLB(lbSummary.DataCenter, lbSummary.LBID)
		}
	}

	return nil, errors.New("requested load balancer was not found by name")	
}

// GetLoadBalancer returns whether the specified load balancer exists, and if so, what its status is.
// NB: status is a single Ingress spec, has nothing to do with operational status
func (clc *CLCCloud) GetLoadBalancer(name, region string) (status *api.LoadBalancerStatus, exists bool, err error) {

	lb,e := findLoadBalancerInstance(clc.clcClient, name,region)
	if e == nil {
		return toStatus(lb.PublicIP), true, nil
	} else {
		return nil, false, e
	}
}

// EnsureLoadBalancer creates a new load balancer 'name', or updates the existing one. Returns the status of the balancer
// For an LB identified by region,name (or created that way, with name=LBID returned) (and possibly desc=serviceName)
//	create a pool for every entry in ports, using serviceAffinity.  Equates api.ServicePort.Port to PoolDetails.IncomingPort
//	for every one of those pools, add a node list from the hosts array
func (clc *CLCCloud) EnsureLoadBalancer(name, region string, 
	loadBalancerIP net.IP, // ignore this.  AWS actually returns error if it's non-nil
	ports []*api.ServicePort, hosts []string, serviceName types.NamespacedName, 
	affinityType api.ServiceAffinity, 
	annotations cloudprovider.ServiceAnnotation) 	(*api.LoadBalancerStatus, error) {

	lb,e := findLoadBalancerInstance(clc.clcClient, name,region)
	if e != nil {	// make a new LB
		inf,e := clc.clcClient.createLB(region,name, serviceName.String())
		if e != nil {
			return nil,e
		}

		lb,e = clc.clcClient.inspectLB(region, inf.LBID)
		if e != nil {
			return nil,e
		}	
	}

	// either way, we now have an LB that answers to (name,region).  Configure it with ports and hosts.
	existingPoolCount := len(lb.Pools)
	desiredPoolCount := len(ports)

	addPorts := make([]api.ServicePort, 0, desiredPoolCount)		// ServicePort specs to make new pools out of
	deletePools := make([]PoolDetails, 0, existingPoolCount)		// unwanted existing PoolDetails to destroy

	fromPorts := make([]api.ServicePort, 0, desiredPoolCount)		// existing port/pool pairs to adjust so they match
	toPools := make([]PoolDetails, 0, desiredPoolCount)

	for _,port := range ports {		
		bMatched := false
		for _,pool := range lb.Pools {
			if port.Port == pool.IncomingPort {   // use ServicePort.Port==PoolDetails.IncomingPort to match
				fromPorts = append(fromPorts, *port)
				toPools = append(toPools, pool)	// insert fromPorts/toPool as a pair only
				bMatched = true
				break
			}
		}

		if !bMatched {
			addPorts = append(addPorts, *port)
		}
	}

	for _,pool := range lb.Pools {
		bMatched := false
		for _,port := range ports {
			if port.Port == pool.IncomingPort {	// would have been sent to fromPorts/toPool above
				bMatched = true
				break
			}
		}

		if !bMatched {
			deletePools = append(deletePools, pool)
		}
	}

	for _,creationPort := range addPorts {
		desiredPool := makePoolDetailsFromServicePort(lb.LBID, &creationPort, hosts, affinityType)
		clc.clcClient.createPool(lb.DataCenter, lb.LBID, desiredPool)
	}

	for _,deletionPool := range deletePools {
		clc.clcClient.deletePool(lb.DataCenter, lb.LBID, deletionPool.PoolID)
	}

	for idx,_ := range fromPorts {
		desiredPort := &fromPorts[idx]	// ServicePort, what K wants
		existingPool := &toPools[idx]	// PoolDetails, what CL has now

		desiredPool := makePoolDetailsFromServicePort(lb.LBID, desiredPort, hosts, affinityType)
		conformPoolDetails(&clc.clcClient, lb.DataCenter, desiredPool, existingPool)
	}

	return toStatus(lb.PublicIP), nil	// ingress is the actual lb.PublicIP, not the one passed in to this func
}

func makePoolDetailsFromServicePort(lbid string, srcPort *api.ServicePort, hosts []string, affinity api.ServiceAffinity) *PoolDetails {
	persist := "none"
	if affinity == "ClientIP" {	// K. calls it this
		persist = "source_ip" // CL calls it that
	}

	return &PoolDetails {
		PoolID:"",	// createPool to fill in
		LBID:lbid,
		IncomingPort:srcPort.Port,
		Method:"roundrobin",
		Persistence:persist,
		TimeoutMS: 99999,	// and what should the default be?
		Mode:"tcp",
		Nodes:makeNodeListFromHosts(hosts, srcPort.NodePort),
	}
}

func conformPoolDetails(clcClient *CenturyLinkClient, dc string, desiredPool, existingPool *PoolDetails) (bool, error) {

	desiredPool.PoolID = existingPool.PoolID
	desiredPool.LBID = existingPool.LBID
	desiredPool.IncomingPort = existingPool.IncomingPort

	bMatch := true 
	if ((desiredPool.Method != existingPool.Method) || (desiredPool.Persistence != existingPool.Persistence)) {
		bMatch = false
	} else if ((desiredPool.TimeoutMS != existingPool.TimeoutMS) || (desiredPool.Mode != existingPool.Mode)) {
		bMatch = false
	} else if (len(desiredPool.Nodes) != len(existingPool.Nodes)) {
		bMatch = false
	} else {
		for idx,_ := range desiredPool.Nodes {
			if desiredPool.Nodes[idx].TargetIP != existingPool.Nodes[idx].TargetIP {
				bMatch = false
			} else if desiredPool.Nodes[idx].TargetPort != existingPool.Nodes[idx].TargetPort {
				bMatch = false
			}
		}
	}

	if bMatch {
		return false,nil	// no changes made, no error
	}

	_,e := (*clcClient).updatePool(dc,desiredPool.LBID, desiredPool)
	return true,e
}


func toStatus(ip string) (*api.LoadBalancerStatus) {
	var ingress api.LoadBalancerIngress
	ingress.Hostname = ip

	ret := api.LoadBalancerStatus{}
	ret.Ingress = []api.LoadBalancerIngress{ingress}

	return &ret
}


// UpdateLoadBalancer updates hosts under the specified load balancer.  For every pool, this rewrites the hosts list.
// We require that every pool must have a nonempty nodes list, deleting pools if necessary to enforce this.
func (clc *CLCCloud) UpdateLoadBalancer(name, region string, hosts []string) error {

	lb,e := findLoadBalancerInstance(clc.clcClient, name,region)
	if e != nil {
		return e	// can't see it?  Can't update it.
	}

	for _,pool := range lb.Pools {

		if ((hosts == nil) || (len(hosts) == 0)) {		// must delete pool
			err := clc.clcClient.deletePool(lb.DataCenter, lb.LBID, pool.PoolID)
			if err != nil {
				return err   // and punt on any other pools.  This LB is in bad shape now.
			}

		} else {	// update hosts in the pool, using port number from the existing hosts

			if ((pool.Nodes == nil) || (len(pool.Nodes) == 0)) {    // no nodes to get targetPort from
				err := clc.clcClient.deletePool(lb.DataCenter, lb.LBID, pool.PoolID)
				if err != nil {
					return err
				}
			} else {	// normal case, draw targetPort from an existing node and rewrite the pool

				targetPort := pool.Nodes[0].TargetPort
				nodelist := makeNodeListFromHosts(hosts, targetPort)

				pool.Nodes = nodelist
				_,err := clc.clcClient.updatePool(lb.DataCenter, lb.LBID, &pool)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func makeNodeListFromHosts(hosts []string, portnum int) ([]PoolNode) {
	nNodes := len(hosts)
	nodelist := make([]PoolNode, nNodes,nNodes)
	for idx,hostnode := range hosts {
		nodelist[idx] = PoolNode {
			TargetIP:hostnode,
			TargetPort:portnum,
		}
	}

	return nodelist
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
// This construction is useful because many cloud providers' load balancers
// have multiple underlying components, meaning a Get could say that the LB
// doesn't exist even if some part of it is still laying around.
func (clc *CLCCloud) EnsureLoadBalancerDeleted(name, region string) (error) {

	lb,e := findLoadBalancerInstance(clc.clcClient, name,region)
	if e == nil {
		_,e = clc.clcClient.deleteLB(lb.DataCenter, lb.LBID)
	}

	return e	// regardless of whether the LB was there previously or not
}

//////////////// Notes about mapping the Kubernetes data model to the CLC LBAAS 
//
// (selecting a particular LB)
// K:name <--> CL:LBID (also CL:LB.LBID)
// K:region <--> CL:datacenter ID
//             CL:accountAlias comes from login creds
//
// (properties of the LB)
// K:IP <--> CL:PublicIP
//             CL:Name (details unknown, needs to be unique)
// K:serviceName <--> CL:Description (probably)
// K:affinity(ClientIP or None) <--> CL:PoolDetails.persistence(source_ip or none)
// K:annotations(ignore)
// K:hosts (array of strings) <--> CL:NodeDetails.IP (same hosts[] for every PoolDetails.Nodes[])
// K:ports (array of ServicePort) <--> CL:Pools (array of PoolDetails)
//
// (properties of a K:ServicePort<-->CL:PoolDetails)
// K:Name <--> CL:PoolID
// K:Protocol(TCP or UDP) (always TCP)
// K:Port <--> CL:Port
// K:TargetPort(ignore)
// K:NodePort <--> CL:NodeDetails.PrivatePort
//              CL:method(roundrobin or leastconn) always roundrobin
//              CL:mode:(tcp or http) always tcp
//              CL:timeout


