// Copyright 2018 Hawkeye Recognition
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
package inventory

import (
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
	deviceGroupUrl = "/api/management/v1/inventory/devices"
)

type Client struct {
	url            string
	deviceGroupUrl string
	client         *http.Client
}

type Device struct {
	Attributes []Attribute
	Id         string
	Updated_ts string
}

type Attribute struct {
	Description string
	Name        string
	Value       string
}

func NewClient(url string, skipVerify bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	return &Client{
		url:            url,
		deviceGroupUrl: JoinURL(url, deviceGroupUrl),
		client: &http.Client{
			Transport: tr,
		},
	}
}

func (c *Client) ListDevices(group string, tokenPath string) ([]Device, error) {
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil, errors.Wrap(err, "Please Login first")
	}
	reqUrl := c.deviceGroupUrl + "?group=" + group
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+string(token))

	reqDump, _ := httputil.DumpRequest(req, true)
	log.Verbf("sending request: \n%v", string(reqDump))

	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "POST /auth/login request failed")
	}
	defer rsp.Body.Close()

	rspDump, _ := httputil.DumpResponse(rsp, true)
	log.Verbf("response: \n%v\n", string(rspDump))

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "can't read request body")
	}

	if rsp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("login failed with status %d", rsp.StatusCode))
	}

	var result []Device
	json.Unmarshal(body, &result)

	return result, nil
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
