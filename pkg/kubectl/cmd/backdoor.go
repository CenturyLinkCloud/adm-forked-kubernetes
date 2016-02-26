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

// Do not submit this file in to the public repo.  The world doesn't want a 'backdoor' command.
// We do, for implementing and testing CLC primitives, especially calls out to the CL API.

// List of files involved in the backdoor:
//  kubectl/cmd/backdoorHTTP.go
//	kubectl/cmd/backdoor.go   (this file, the command itself)
//  kubectl/cmd/cmd.go   (one-line change, to register this command in the dispatcher)

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)


type BackdoorOptions struct {
	// and how do the options get populated?
	Arguments []string
}

const (
	backdoor_long = `Shadow API used for testing access to CLC primitives.  Do not submit to public repo.
`

	backdoor_example = `Example usage goes here
`
)

func NewCmdBackdoor(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	//	options := &BackdoorOptions{}

	cmd := &cobra.Command{
		Use:     "backdoor (any string sequence)",
		Short:   "direct access to CLC primitives",
		Long:    backdoor_long,
		Example: backdoor_example,
		Run: func(cmd *cobra.Command, args []string) {
			err := RunBackdoor(f, out, args)
			cmdutil.CheckErr(err)
		},
	}

	return cmd
}

//  
//  // make a new client, and auth it
//  func GetClientCLC() (*clc_sdk.Client,error) {		
//  	AuthCLC,err := clc_sdk_api.EnvConfig()
//  	if err != nil {
//  		fmt.Printf("CLC backdoor could not load auth config\n");
//  		return nil,err
//  	} else {
//  		fmt.Printf("CLC API backdoor access, auth alias: %s\n", AuthCLC.Alias)
//  	}
//  
//  	ClientCLC := clc_sdk.New(AuthCLC)  // no error return
//  	err2 := ClientCLC.Authenticate()
//  	if err2 != nil {
//  		fmt.Printf("CLC backdoor could not authenticate\n")
//  		return nil,err2;
//  	} else {
//  		fmt.Printf("CLC authenticated\n")
//  	}
//  
//  	return ClientCLC,nil
//  }
//  
//  func TestGetDataCenterList(client *clc_sdk.Client) error {
//  	AllDC,err := client.DC.GetAll()
//  	if(err != nil) {
//  		fmt.Printf("CLC: could not get list of data centers\n")
//  		return err;
//  	}
//  
//  	if AllDC == nil {
//  		fmt.Printf("CLC, null list of data centers is probably an error\n")
//  	}
//  
//  	fmt.Printf("CLC: have the data center list successfully\n");
//  
//  	for _,val := range AllDC {
//  		fmt.Printf("CLC DC: %s %s\n", val.ID, val.Name)
//  	}
//  
//  	return err
//  }
//  
//  func TestCreateLoadBalancer(client *clc_sdk.Client) (*clc_sdk_lb.LoadBalancer,error) {
//  	newLB := clc_sdk_lb.LoadBalancer {
//  		Name: "test LB",
//  		Description: "testing golang SDK",
//  	}
//  
//  	resp,err := client.LB.Create("wa1", newLB)
//  	if(err != nil) {
//  		fmt.Printf("CLC: failed to create load balancer\n");
//  		return resp,err
//  	} else {
//  		fmt.Printf("CLC: made a load balancer, ID=%s, addr=%s \n", resp.ID, resp.IPaddress)
//  		return resp,err
//  	}
//  }
//  

func printUsage() {
	fmt.Printf("==== Kubectl usage: ====\n")
	fmt.Printf("kubectl backdoor help \n")
	fmt.Printf("kubectl backdoor auth login <username> <password> \n")
	fmt.Printf("kubectl backdoor auth logout \n")
	fmt.Printf("kubectl backdoor DC list \n")
	fmt.Printf("kubectl backdoor LB create dc lbName lbDesc \n")
	fmt.Printf("kubectl backdoor LB delete dc lbID \n")
	fmt.Printf("kubectl backdoor LB details dc lbID \n")
	fmt.Printf("kubectl backdoor LB list \n")
	fmt.Printf("kubectl backdoor pool create dc lbID \n")
	fmt.Printf("kubectl backdoor pool update dc lbID poolID \n")
	fmt.Printf("kubectl backdoor pool delete dc lbID poolID \n")
}

func cmdHelp(args []string) {
	// nyi help on specific commands
	printUsage();
}

// 'backdoor' not passed in, i.e. "kubectl backdoor foo bar baz" gives three args: foo, bar, baz
func RunBackdoor(f *cmdutil.Factory, out io.Writer, args []string) error {
	if (len(args) == 0) {
		printUsage();
	} else if((len(args) >= 1) && (args[0] == "help")) {							// backdoor help
		cmdHelp(args);
	} else if ((len(args) >= 2) && (args[0] == "auth") && (args[1] == "login")) {	// backdoor auth login
		cmdAuthLogin(args);
	} else if ((len(args) >= 2) && (args[0] == "auth") && (args[1] == "logout")) {	// backdoor auth logout
		cmdAuthLogout(args);
	} else if ((len(args) == 2) && (args[0] == "DC") && (args[1] == "list")) {      // backdoor DC list
		cmdListAllDC()
	} else if ((len(args) == 5) && (args[0] == "LB") && (args[1] == "create")) {	// backdoor LB create dc lbName lbDesc
		cmdLoadBalancerCreate(args[2], args[3], args[4]);
	} else if ((len(args) == 4) && (args[0] == "LB") && (args[1] == "delete")) {	// backdoor LB delete dc LBID
		cmdLoadBalancerDelete(args[2], args[3])
	} else if ((len(args) >= 2) && (args[0] == "LB") && (args[1] == "details")) {	// backdoor LB details dc LBID
		cmdLoadBalancerDetails(args[2], args[3])
	} else if ((len(args) == 2) && (args[0] == "LB") && (args[1] == "list")) {		// backdoor LB list
		cmdListAllLB()
	} else if ((len(args) == 4) && (args[0] == "pool") && (args[1] == "create")) {	// backdoor pool create dc lbid
		cmdCreatePool(args[2], args[3])
	} else if ((len(args) == 5) && (args[0] == "pool") && (args[1] == "update")) {	// backdoor pool update dc lbid poolID
		cmdUpdatePool(args[2], args[3], args[4])
	} else if ((len(args) == 5) && (args[0] == "pool") && (args[1] == "delete")) {	// backdoor pool delete dc lbid poolID
		cmdDeletePool(args[2], args[3], args[4])
	} else {
		fmt.Printf("could not parse command\n");
		printUsage();
	}

	return nil
}


// api inputs:   GET, POST(json struct), POST(string)
//		bDebugRequest, bDebugResponse, bConnectionClose
//		URI (relative to API stem)

// response: error relating to setting up the request (incl. not having previous auth)
//		could not execute http call
//		HTTP status code (even in success)
//		response body (parsed into JSON struct?)


func cmdAuthLogin(args []string) {
	if (len(args) >= 4) {		// auth login <username> <password> [close | dumprequest | dumpresponse]*
		bCloseConnection := false
		bDumpRequest := false
		bDumpResponse := false
		for idxArg := 4 ; idxArg < len(args) ; idxArg++ {
			if args[idxArg] == "close" {
				bCloseConnection = true
			} else if args[idxArg] == "dump-request" {
				bDumpRequest = true
			} else if args[idxArg] == "dump-response" {
				bDumpResponse = true
			} else {
				fmt.Printf("unrecognized argument %s\n", args[idxArg])
			}
		}

		clc,err := ClientLogin(args[2], args[3])
		if(err != nil) {
			fmt.Println(err)
			return
		}

	// suppress compiler annoyance
	bCloseConnection = bDumpRequest;
	bDumpRequest = bDumpResponse;
	bDumpResponse = bCloseConnection;

		fmt.Printf("successful login, alias=%s\n", clc.getAccountAlias())
	} else {
		fmt.Printf("auth login:  invalid arguments\n");
		printUsage();
	}
}

func cmdAuthLogout(args []string) {
	fmt.Printf("no true logout inside kubectl-backdoor, consider export CLC_API_TOKEN=\n")
}

func cmdListAllDC() {
	clc, err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	dclist,err := clc.listAllDC()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _,dc := range dclist {
		fmt.Printf("DC: id=%s, name=\"%s\"\n", dc.DCID, dc.Name)
	}
}

func cmdListAllLB() {
	clc, err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	lblist,err := clc.listAllLB()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _,lb := range lblist {	// we get LBSummary back
		fmt.Printf("LB: id=%s, name=\"%s\", desc=\"%s\",\n    ip=%s, dc=%s\n",
			lb.LBID, lb.Name, lb.Description, lb.PublicIP, lb.DataCenter)
	}
}

func cmdLoadBalancerCreate(argDC, argName, argDesc string) {
	clc,err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	lbinf,err := clc.createLB(argDC, argName, argDesc)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("createLB status: time=%d id=%s \n", lbinf.RequestTime, lbinf.LBID)
}

func cmdLoadBalancerDelete(argDC, argLBID string) {
	clc,err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = clc.deleteLB(argDC, argLBID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("load balancer deleted\n")
}


func cmdLoadBalancerDetails(argDC, argLBID string) {
	clc,err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	lb,err := clc.inspectLB(argDC, argLBID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Queried LB: id=%s, status=%s, name=%s, description=%s, dc=%s, IP=%s \n",
		lb.LBID, lb.Status, lb.Name, lb.Description, lb.DataCenter, lb.PublicIP)
}

func cmdCreatePool(argDC, argLBID string) {
	clc,err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	newpoolinfo := PoolDetails{	// dummy info at this point
		LBID:argLBID,
		IncomingPort:80,

		Method:"roundrobin",	// CLC API makes these mandatory
		Mode:"http",
		Persistence:"none",
		TimeoutMS:10,
	}

	pool,err := clc.createPool(argDC, argLBID, &newpoolinfo)
	if err != nil {
		fmt.Println(err)
		return
	}

	printPoolDetails(pool)
}

func printPoolDetails(pool *PoolDetails) {
	fmt.Printf("Pool: poolID=%s LBID=%s port=%d, method=%s persistence=%s mode=%s \n",
		pool.PoolID, pool.LBID, pool.IncomingPort, pool.Method, pool.Persistence, pool.Mode)

	for _,n := range pool.Nodes {
		fmt.Printf("  pool node:  IP=%s port=%d \n", n.TargetIP, n.TargetPort)
	}
}

func cmdUpdatePool(argDC, argLBID, argPOOLID string) {
	clc,err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	newpoolinfo := PoolDetails{	// dummy info at this point
		PoolID:argPOOLID,
		LBID:argLBID,
		IncomingPort:8080,

		Method:"leastconn",	// CLC API makes these mandatory
		Mode:"tcp",
		Persistence:"none",
		TimeoutMS:10,
	}

	pool,err := clc.updatePool(argDC,argLBID, &newpoolinfo)
	if err != nil {
		fmt.Println(err)
		return
	}

	printPoolDetails(pool)
}

func cmdDeletePool(argDC, argLBID, argPOOLID string) {
	clc,err := ClientReload()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = clc.deletePool(argDC, argLBID, argPOOLID)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("pool deleted\n")
}

