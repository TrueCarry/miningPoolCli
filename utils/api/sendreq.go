/*
miningPoolCli – open-source tonuniverse mining pool client

Copyright (C) 2021 tonuniverse.com

This file is part of miningPoolCli.

miningPoolCli is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

miningPoolCli is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with miningPoolCli.  If not, see <https://www.gnu.org/licenses/>.
*/

package api

import (
	"bytes"
	"crypto/tls"
	"miningPoolCli/config"
	"miningPoolCli/utils/mlog"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

type ServerResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
	Code   int    `json:"code"`
}

var (
	proxyClient = fasthttp.Client{
		MaxConnsPerHost: 4,

		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,

		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
)

func sendPostJsonReqAttempt(jsonData []byte, serverUrl string) ([]byte, error) {
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(serverUrl)
	httpReq.Header.SetMethod(fasthttp.MethodPost)
	httpReq.Header.SetContentType("application/json; charset=UTF-8")
	httpReq.Header.Set("Build-Version", config.BuildVersion)
	httpReq.SetBody(jsonData)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	err := proxyClient.DoTimeout(httpReq, httpResp, 5*time.Second)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = httpResp.BodyWriteTo(&buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func GetReqAttempt(serverUrl string) ([]byte, error) {
	httpReq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(httpReq)

	httpReq.SetRequestURI(serverUrl)
	httpReq.Header.SetMethod(fasthttp.MethodGet)
	httpReq.Header.SetContentType("application/json; charset=UTF-8")
	httpReq.Header.Set("Build-Version", config.BuildVersion)

	httpResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(httpResp)

	err := proxyClient.DoTimeout(httpReq, httpResp, 5*time.Second)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = httpResp.BodyWriteTo(&buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func SendPostJsonReq(jsonData []byte, serverUrl string) []byte {
	var body []byte = nil
	for attempts := 0; attempts < 5; attempts++ {
		var err error
		body, err = sendPostJsonReqAttempt(jsonData, serverUrl)
		if err != nil {
			mlog.LogError(err.Error())
			mlog.LogInfo("Sleep request for 3 sec")
			time.Sleep(3 * time.Second)
			mlog.LogInfo("Attempting to retry the request... [" + strconv.Itoa(attempts+1) + "/" + "3]")
			continue
		}

		if attempts > 0 {
			mlog.LogOk("Request sent")
		}

		break
	}
	// if body == nil {
	// 	mlog.LogFatal("Attempts to send a request have yielded no results :(")
	// }

	return body
}
