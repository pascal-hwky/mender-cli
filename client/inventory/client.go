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
	"net/http"
	"strings"
)

const (
	deviceGroupUrl = "/api/management/v1/inventory/devices"
)

type Client struct {
	url            string
	deviceGroupUrl string
	client         *http.Client
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

func (c *Client) ListDevices(group string) ([]byte, error) {
	return nil, nil
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
