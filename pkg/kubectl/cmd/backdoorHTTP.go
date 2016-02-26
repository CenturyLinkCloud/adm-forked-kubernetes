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
	"errors"
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"encoding/json"
	tls "crypto/tls"
)

//// requests honor this state, no need to pass in with every call
var bCloseConnections = true;
var bDebugRequests = true;
var bDebugResponses = true;

func SetCloseConnectionMode(b bool) {
	bCloseConnections = b;
}

func SetDebugRequestMode(b bool) {
	bDebugRequests = b;
}

func SetDebugResponseMode(b bool) {
	bDebugResponses = b;
}


//// Credentials is returned from the login func, and used by everything else
type Credentials struct {
	Username string 
	AccountAlias string
	LocationAlias string  // do we need this?
	BearerToken string
}

func (obj *Credentials) GetUsername() (string) {
	return obj.Username
}

func (obj *Credentials) GetAccount() (string) {
	return obj.AccountAlias
}

func (obj *Credentials) GetLocation() (string) {
	return obj.LocationAlias
}

func (obj *Credentials) IsValid() bool {
	return (obj.AccountAlias != "") && (obj.BearerToken != "")
}

// and no GetBearerToken - it's private within this file

func (obj *Credentials) ClearCredentials() {	// creds object is useless after this
	obj.Username = ""
	obj.AccountAlias = ""
	obj.LocationAlias = ""
	obj.BearerToken = ""
}


func makeError(content string) error {
	if(content == "") { 
		content = "<error text not available>"
	}

	return errors.New("CLC API: " + content)
}

var dummyCreds = Credentials{ Username:"dummy object passed by login proc and not used",
	AccountAlias:"invalid", LocationAlias:"invalid", BearerToken:"invalid", }	// note dummyCreds.IsValid() is true

func GetCredentials(server, uri string, username, password string) (*Credentials, error) {
	if (username == "") || (password == "") {
		return nil, makeError("username and/or password not available")
	}

	body := fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", username,password);
	b := bytes.NewBufferString(body)

	authresp := AuthLoginResponseJSON { }

	err := invokeHTTP("POST", server, uri, &dummyCreds, b, &authresp)
	if err != nil {
		return nil,err
	}

	fmt.Printf("assigning new token, do this:  export CLC_API_TOKEN=%s\n", authresp.BearerToken)
	fmt.Printf("also CLC_API_USERNAME=%s  CLC_API_ACCOUNT=%s  CLC_API_LOCATION=%s\n", authresp.Username, authresp.AccountAlias, authresp.LocationAlias)

	return &Credentials {
		Username: authresp.Username ,
		AccountAlias: authresp.AccountAlias ,
		LocationAlias: authresp.LocationAlias ,
		BearerToken: authresp.BearerToken ,
		}, nil
}

type AuthLoginRequestJSON struct {	// actually this is unused, as we simply sprintf the string
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthLoginResponseJSON struct {
	Username string `json:"username"`
	AccountAlias string `json:"accountAlias"`
	LocationAlias string `json:"locationAlias"`
	Roles []string `json:"roles"`
	BearerToken string `json:"bearerToken"`
}


// no request message body sent.  Response body returned if ret is not nil
func simpleGET(server, uri string, creds *Credentials, ret interface{} ) error {
	return invokeHTTP("GET", server, uri, creds, nil, ret)
}

// no request message body sent.  Respone body returned if ret is not nil
func simpleDELETE(server, uri string, creds *Credentials, ret interface{} ) error {
	return invokeHTTP("DELETE", server, uri, creds, nil, ret)
}

// body must be a json-annotated struct, and is marshalled into the request body
func marshalledPOST(server, uri string, creds *Credentials, body interface{}, ret interface{} ) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(body)
	if err != nil {
		return err		// JSON marshalling failed
	}

	return invokeHTTP("POST", server, uri, creds, b, ret)
}

// body is a JSON string, sent directly as the request body
func simplePOST(server, uri string, creds *Credentials, body string, ret interface{} ) error {
	b := bytes.NewBufferString(body)
	return invokeHTTP("POST", server, uri, creds, b, ret)
}


// method to be "GET", "POST", etc.
// server name "api.ctl.io" or api.loadbalancer.ctl.io
// uri always starts with /   (we assemble https://<server><uri>)
// creds required for anything except the login call
// body may be be nil
func invokeHTTP(method, server, uri string, creds *Credentials, body io.Reader, ret interface{} ) error {
	if (creds == nil) || !creds.IsValid() {
		return makeError("no credentials provided")
	}

	req,err := http.NewRequest(method, ("https://" + server + uri), body)
	if err != nil {
		return err
	} else if body != nil {
		req.Header.Add("Content-Type", "application/json")	// incoming body to be a marshaled object already
	}

	req.Header.Add("Host", server)	// the reason we take server and uri separately
	req.Header.Add("Accept", "application/json")

	if(creds != &dummyCreds) {	// the login proc itself doesn't send an auth header
		req.Header.Add("Authorization", ("Bearer " + creds.BearerToken))
	}

	if bCloseConnections {
		req.Header.Add("Connection", "close")
	}

	if bDebugRequests {
		v, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(v))
	}


// this should be the normal code
//	resp,err := http.DefaultClient.Do(req)	// execute the call

// instead, we have this which tolerates bad certs
	tlscfg := &tls.Config { InsecureSkipVerify: true }    // true means to skip the verification
	transp := &http.Transport { TLSClientConfig: tlscfg }
	client := &http.Client{ Transport:transp }
	resp,err := client.Do(req)
// end of tolerating bad certs.  Do not keep this code - it allows MITM etc. attacks


	if bDebugResponses {
		vv,_ := httputil.DumpResponse(resp,true)
		fmt.Println(string(vv))
	}

	if err != nil {   // failed HTTP call
		return err
	}

	if ((resp.StatusCode < 200) || (resp.StatusCode >= 300)) {	// Q: do we care to distinguish the various 200-series codes?
		stat := fmt.Sprintf("received HTTP response code %d\n", resp.StatusCode)

		if !bDebugRequests {
			fmt.Printf("dumping this request, after the fact\n")
			v, _ := httputil.DumpRequestOut(req,true)
			fmt.Println(string(v))
		}

		if !bDebugResponses {
			vv,_ := httputil.DumpResponse(resp,true)
			fmt.Println(string(vv))
		}

		return makeError(stat)
	}

	if ret != nil {	// permit methods without a response body, or calls that ignore the body and just look for status
		err = json.NewDecoder(resp.Body).Decode(ret)
	}

	return err
}

