// Copyright 2018 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package deploy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/mendersoftware/mender-cli/log"
	"github.com/pkg/errors"
)

const (
	deployUrl = "/api/management/v1/deployments/deployments"
)

type Client struct {
	url       string
	deployUrl string
	client    *http.Client
}

func NewClient(url string, skipVerify bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	return &Client{
		url:       url,
		deployUrl: JoinURL(url, deployUrl),
		client: &http.Client{
			Transport: tr,
		},
	}
}

func (c *Client) DeployRelease(artifactName string, deviceIds []string, deployName string, tokenPath string) error {
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return errors.Wrap(err, "Please Login first")
	}
	reqUrl := c.deployUrl
	deviceIdsJson, _ := json.Marshal(deviceIds)
	var jsonStr = []byte(`{"artifact_name": "` + artifactName + `", "devices": ` + string(deviceIdsJson) + `, "name": "` + deployName + `"}`)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	reqDump, _ := httputil.DumpRequest(req, true)
	log.Verbf("sending request: \n%v", string(reqDump))

	rsp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "POST /auth/login request failed")
	}
	defer rsp.Body.Close()

	rspDump, _ := httputil.DumpResponse(rsp, true)
	log.Verbf("response: \n%v\n", string(rspDump))

	//body, err := ioutil.ReadAll(rsp.Body)
	//if err != nil {
	//		return errors.Wrap(err, "can't read request body")
	//}

	if rsp.StatusCode != http.StatusCreated {
		return errors.New(fmt.Sprintf("deploy failed with status %d", rsp.StatusCode))
	}

	return nil
}

func JoinURL(base, url string) string {
	if strings.HasPrefix(url, "/") {
		url = url[1:]
	}
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}
	return base + url
}
