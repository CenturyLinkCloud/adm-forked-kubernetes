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

package cmd

import (
	"fmt"
	"os"
	"strings"
)

//// api use involves calls to both addresses
var clcServer_API_V2 string = "api.ctl.io"   // URL form  https://api.ctl.io/v2/<resource>/<accountAlias>
var clcServer_LB_BETA string = "api.loadbalancer.ctl.io"   // URL form  https://api.loadbalancer.ctl.io/<accountAlias>/<datacenter>/loadbalancers


//// auth methods
func implClientLogin(username, password string) (CenturyLinkClient, error) {

	newcreds,err := GetCredentials(clcServer_API_V2, "/v2/authentication/login", username,password)
	if(err != nil) {
		return nil,err
	}

	return clcImpl { 
		creds:newcreds, 
	}, nil
}

func implClientFromEnv() (CenturyLinkClient, error) {

	envUsername := os.Getenv("CLC_API_USERNAME")
	envAccount := os.Getenv("CLC_API_ACCOUNT")
	envLocation := os.Getenv("CLC_API_LOCATION")
	envToken := os.Getenv("CLC_API_TOKEN")

	if ((envUsername == "") || (envAccount == "") || (envLocation == "") || (envToken == "")) {
		return nil, makeError("CLC auth not set in env")
	}

	newcreds := &Credentials { Username:envUsername, AccountAlias:envAccount, LocationAlias:envLocation, BearerToken:envToken, }
	return clcImpl{ creds:newcreds, } , nil
}

//// clcImpl is the internal layer that knows what HTTP calls to make


type clcImpl struct {		// implements CenturyLinkClient
	creds *Credentials
}

func (clc clcImpl) logout() {
	if clc.creds != nil {
		clc.creds.ClearCredentials();
		clc.creds = nil
	}
}

func (clc clcImpl) hasCredentials() bool {
	if clc.creds != nil {
		return clc.creds.IsValid()
	}

	return false
}

// inconsistent style - are methods supposed to start with capital or not?
func (clc clcImpl) getUsername() string {
	if clc.creds != nil {
		return clc.creds.GetUsername()
	}

	return ""
}

func (clc clcImpl) getAccountAlias() string {
	if clc.creds != nil {
		return clc.creds.GetAccount()
	}

	return ""
}

//////////////// clc method: listAllDC()

type dcNamesJSON struct {
	ID string `json:"id"`
	Name string `json:"name"`		// NB: capitalizing Name is required in Go
}

func (clc clcImpl) listAllDC() ([]DataCenterName, error) {

	uri := fmt.Sprintf("/v2/datacenters/%s", clc.creds.GetAccount())
	dcret := make([]*dcNamesJSON, 0)

	err := simpleGET(clcServer_API_V2, uri, clc.creds, &dcret)
	if err != nil {
		return nil, err
	}

	ret := make([]DataCenterName, len(dcret))
	for idx, dc := range dcret {
		ret[idx] = DataCenterName{ DCID:strings.ToUpper(dc.ID), Name:dc.Name }	// api returns ID in lowercase but requires it in upper
	}

	return ret,nil
}


//////////////// clc method: listAllLB()

type lbListingDetailsJSON struct {
	LBID string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	PublicIP string `json:"publicIPAddress"`
//	PrivateIP string `json:"privateIPAddress"`
//	Pools []lbPoolDetailsJSON // ??json format?
//	Status string `json:"status"`
//	AccountAlias string `json:"accountAlias"`
	DataCenter string `json:"dataCenter"`
	// omit keepalivedRouterId and version
}

type lbListingWrapperJSON struct {
	Values []lbListingDetailsJSON `json:"values"`	
}

func (clc clcImpl) listAllLB() ([]LoadBalancerSummary, error) {

	uri := fmt.Sprintf("/%s/loadbalancers", clc.creds.GetAccount())
	apiret := &lbListingWrapperJSON { }

	err := simpleGET(clcServer_LB_BETA, uri, clc.creds, &apiret)
	if(err != nil) {
		return nil,err
	}

	ret := make([]LoadBalancerSummary, len(apiret.Values))
	for idx,lb := range apiret.Values {
		ret[idx] = LoadBalancerSummary {
			LBID: lb.LBID,
			Name: lb.Name,
			Description: lb.Description,
			PublicIP: lb.PublicIP,
			DataCenter: lb.DataCenter,
		}
	}

	return ret,nil
}


//////////////// clc method: createLB()

type LinkJSON struct {
	Rel string `json:"rel,omitempty"`
	Href string `json:"href,omitempty"`
	ID string `json:"resourceId,omitempty"`
}

type ApiLinks []LinkJSON


type lbCreateRequestJSON struct {
	LBID string `json:"id"`
	Status string `json:"status"`
	Description string `json:"description"`
	RequestDate int64 `json:"requestDate"`
	CompletionDate int64 `json:"completionDate"`
	Links ApiLinks `json:"links"`
}

func (clc clcImpl) createLB(dc string, lbname string, desc string) (*LoadBalancerCreationInfo, error) {

	uri := fmt.Sprintf("/%s/%s/loadbalancers", clc.creds.GetAccount(), dc)
	apiret := &lbCreateRequestJSON{ }
	
	body := fmt.Sprintf("{ \"name\":\"%s\", \"description\":\"%s\" }", lbname, desc)

	err := simplePOST(clcServer_LB_BETA, uri, clc.creds, body, apiret)

	if err != nil {
		return nil,err
	}

	return &LoadBalancerCreationInfo {
		LBID: findLinkLB(&apiret.Links, "loadbalancer"),
		RequestTime: apiret.RequestDate,
	}, nil
}


//////////////// clc method: inspectLB()

type NodeJSON struct {
	TargetIP string `json:"ipAddress"`
	TargetPort int `json:"privatePort"`
}

type ApiNodes []NodeJSON

type PoolJSON struct {
	PoolID string `json:"id"`
	IncomingPort int `json:"port"`
	Method string `json:"loadBalancingMethod"`
	Persistence string `json:"persistence"`
	TimeoutMS int64 `json:"idleTimeout"`
	Mode string `json:"loadBalancingMode"`
	// healthCheck not modeled

	Nodes ApiNodes`json:"nodes"`
}

type ApiPools []PoolJSON

type lbDetailsJSON struct {
	LBID string `json:"id"`
	Status string `json:"status"`
	Name string `json:"name"`
	Description string `json:"description"`
	PublicIP string `json:"publicIPAddress"`
	DataCenter string `json:"dataCenter"`
	Pools ApiPools `json:"pools"`
}

func nodes_from_json(src *ApiNodes) (*[]PoolNode) {
	nNodes := len(*src)
	if nNodes == 0 {
		return nil
	}

	ret := make([]PoolNode, nNodes, nNodes)
	for idx,srcnode := range *src {
		ret[idx] = PoolNode{
			TargetIP:srcnode.TargetIP,
			TargetPort:srcnode.TargetPort,
		}
	}

	return &ret
}

func pools_from_json(src *ApiPools, lbid string) (*[]PoolDetails) {
	nPools := len(*src)
	if nPools == 0 {
		return nil
	}

	ret := make([]PoolDetails, nPools, nPools)
	for idx,srcpool := range *src {
		ret[idx] = PoolDetails{
			PoolID:srcpool.PoolID,
			LBID:lbid,
			IncomingPort:srcpool.IncomingPort,
			Method:srcpool.Method,
			Persistence:srcpool.Persistence,
			TimeoutMS:srcpool.TimeoutMS,
			Mode:srcpool.Mode,
			Nodes: *nodes_from_json(&srcpool.Nodes),
		}
	}

	return &ret
}

func (clc clcImpl) inspectLB(dc, lbid string) (*LoadBalancerDetails, error) {

	uri := fmt.Sprintf("/%s/%s/loadbalancers/%s", clc.creds.GetAccount(), dc, lbid)
	apiret := &lbDetailsJSON { }

	err := simpleGET(clcServer_LB_BETA, uri, clc.creds, apiret)
	if err != nil {
		return nil,err
	}

	return &LoadBalancerDetails {
		LBID: apiret.LBID,
		Status: apiret.Status,
		Name: apiret.Name,
		Description: apiret.Description,
		PublicIP: apiret.PublicIP,
		DataCenter: apiret.DataCenter,
		Pools: *pools_from_json(&apiret.Pools, apiret.LBID),
		}, nil
}


//////////////// clc method: deleteLB()


// unneeded, since now deleteLB doesn't return any of this
type lbDeleteJSON struct {
	ID string `json:"id"`
	Status string `json:"status"`
	Description string `json:"description"`
	RequestTime int64 `json:"requestDate"`
	CompletionTime int64 `json:"completionTime"`
	Links ApiLinks `json:"links"`
}

func findLinkLB(links *ApiLinks, rel string) string {
	for _,link := range *links {
		if link.Rel == rel {
			return link.ID
		}
	}

	return ""  // not found, consider returning err?
}

func (clc clcImpl) deleteLB(dc, lbid string) (error) {
	
	uri := fmt.Sprintf("/%s/%s/loadbalancers/%s", clc.creds.GetAccount(), dc, lbid)
	apiret := &lbDeleteJSON { }

	err := simpleDELETE(clcServer_LB_BETA, uri, clc.creds, apiret)
	return err
}


//////////////// clc method: createPool()

type NodeEntityJSON struct {
	TargetIP string `json:"ipAddress"`
	TargetPort int `json:"privatePort"`
}

type PoolEntityJSON struct {
	PoolID string `json:"id,omitempty"`		// createPool doesn't want to send an id
	IncomingPort int `json:"port"`
	Method string `json:"loadBalancingMethod"`
	Persistence string `json:"persistence"`
	TimeoutMS int64 `json:"idleTimeout"`
	Mode string `json:"loadBalancingMode"`
	Nodes []NodeEntityJSON `json:"nodes"`
}

func pool_to_json(pool *PoolDetails) *PoolEntityJSON {
	return &PoolEntityJSON{
		PoolID: pool.PoolID,
		IncomingPort: pool.IncomingPort,
		Method: pool.Method,
		Persistence: pool.Persistence,
		TimeoutMS: pool.TimeoutMS,
		Mode: pool.Mode,
		Nodes: *nodes_to_json(&pool.Nodes),
	}
}

func nodes_to_json(src *[]PoolNode) (*[]NodeEntityJSON) {
	nNodes := len(*src)
	if nNodes == 0 {
		return nil
	}

	ret := make([]NodeEntityJSON, nNodes, nNodes)
	for idx,srcnode := range *src {
		ret[idx] = NodeEntityJSON{
			TargetIP:srcnode.TargetIP,
			TargetPort:srcnode.TargetPort,
		}
	}

	return &ret
}



type CreatePoolResponseJSON struct {
	RequestID string `json:"id"`	
	Status string `json:"status"`
	Description string `json:"description"`
	RequestDate int64 `json:"requestDate"`
	CompletionDate int64 `json:"completionDate"`
	Links ApiLinks `json:"links"`	// Q: does the marshaling work if all we include is this one field?  All we need is links[rel="pool"].resourceID
}

func (clc clcImpl) createPool(dc,lbid string, newpool *PoolDetails) (*PoolDetails, error) {

	uri := fmt.Sprintf("/%s/%s/loadbalancers/%s/pools", clc.creds.GetAccount(), dc, lbid)
	pool_req := pool_to_json(newpool)

	pool_resp := &CreatePoolResponseJSON { }
	err := marshalledPOST(clcServer_LB_BETA, uri, clc.creds, pool_req, pool_resp)
	if err != nil {
		return nil,err
	}

	poolID := findLinkLB(&pool_resp.Links, "pool")
	if(poolID == "") {
		return nil,makeError("could not determine ID of new pool")
	}

	return clc.inspectPool(dc,lbid, poolID)
}


//////////////// clc method: updatePool()
func (clc clcImpl) updatePool(dc,lbid string, newpool *PoolDetails) (*PoolDetails, error) {

	uri := fmt.Sprintf("/%s/%s/loadbalancers/%s/pools/%s", clc.creds.GetAccount(),
		dc, lbid, newpool.PoolID)

	update_req := pool_to_json(newpool)	// and ignore async-request return object
	err := marshalledPOST(clcServer_LB_BETA, uri, clc.creds, update_req, nil)
	if err != nil {
		return nil,err
	}

	return clc.inspectPool(dc,lbid, newpool.PoolID)
}


//////////////// clc method: deletePool()
func (clc clcImpl) deletePool(dc,lbid string, poolID string) error {

	uri := fmt.Sprintf("/%s/%s/loadbalancers/%s/pools/%s", clc.creds.GetAccount(),
		dc, lbid, poolID)

	err := simpleDELETE(clcServer_LB_BETA, uri, clc.creds, nil)
	return err		// no other return body
}


//////////////// clc method: inspectPool()
// not actually part of the LBAAS interface at this time.  Synthesized by returning just part of the inspectLB response

func (clc clcImpl) inspectPool(dc,lbid,poolid string) (*PoolDetails, error) {

	lbDetails,err := clc.inspectLB(dc,lbid)
	if(err != nil) {
		return nil,err
	}

	for _,p := range lbDetails.Pools {
		if p.PoolID == poolid {
			return &p,nil		// success
		}
	}

	return nil,makeError("pool not found")
}

//////////////// end clc methods
