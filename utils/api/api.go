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
	"encoding/json"
	"errors"
	"miningPoolCli/config"
	"miningPoolCli/utils/mlog"
)

type User struct {
	Id      int    `json:"id"`
	Address string `json:"address"`
	Token   string `json:"token"`
	Balance int    `json:"balance"`
}

type AuthResponse struct {
	User        User   `json:"user"`
	PoolAddress string `json:"pool_address"`
	ServerResponse
}

func Auth() {
	jsonData, _ := json.Marshal(map[string]string{"token": config.ServerSettings.AuthKey})
	bodyResp := SendPostJsonReq(
		jsonData,
		config.ServerSettings.MiningPoolServerURL+"/token",
	)

	var serverResp AuthResponse

	err := json.Unmarshal(bodyResp, &serverResp)
	if err != nil {
		mlog.LogFatalStackError(err)
	}

	if serverResp.User.Id != 0 {
		mlog.LogOk("Authorization successful\n")
		if serverResp.User.Address != "" {
			mlog.LogInfo("Your TON wallet:")
			mlog.LogInfo(serverResp.User.Address)
		} else {
			mlog.LogInfo("You can set your TON wallet in https://ton.ninja")
		}

		config.StaticBeforeMinerSettings.PoolAddress = serverResp.PoolAddress
	} else {
		mlog.LogFatal("Auth failed; invalid token")
	}
	mlog.LogPass()
}

type Task struct {
	Id         int    `json:"id"`
	Seed       string `json:"seed"`
	Complexity string `json:"new_complexity"`
	Giver      string `json:"address"`
	Expire     int64  `json:"expire"`
}

type GetTasksResponse struct {
	Tasks []Task `json:"tasks"`
	ServerResponse
}

func GetTasks() (GetTasksResponse, error) {
	jsonData, _ := json.Marshal(map[string]string{})
	bodyResp := SendPostJsonReq(
		jsonData,
		config.ServerSettings.MiningPoolServerURL+"/get",
	)

	if bodyResp == nil {
		return GetTasksResponse{}, errors.New("no resp")
	}

	var results GetTasksResponse

	if err := json.Unmarshal(bodyResp, &results); err != nil {
		mlog.LogError(err.Error())
		mlog.LogError("can not unmarshal JSON GetTasks()")
		mlog.LogError("bodyResp: " + string(bodyResp))
	}

	return results, nil
}

type SendHexBocToServerResponse struct {
	ServerResponse
	Hash       string `json:"hash"`
	Complexity string `json:"complexity"`
}

func SendHexBocToServer(hexData string, seed string, taskId string) (SendHexBocToServerResponse, error) {
	jsonData, _ := json.Marshal(map[string]string{
		"hexData":    hexData,
		"dataSource": "minerClient",
		"token":      config.ServerSettings.AuthKey,
		"speed":      "1",
		"seed":       seed,
		"id":         taskId,
	})

	bodyResp := SendPostJsonReq(
		jsonData,
		config.ServerSettings.MiningPoolServerURL+"/boc",
	)

	var results SendHexBocToServerResponse
	if bodyResp == nil {
		return results, errors.New("no response body")
	}

	if err := json.Unmarshal(bodyResp, &results); err != nil {
		mlog.LogError(err.Error())
		mlog.LogError("Can not unmarshal JSON SendHexBocToServer()")
		mlog.LogError("bodyResp: " + string(bodyResp))

		return results, errors.New("can not unmarshal json SendHexBocToServer()")
	}

	return results, nil
}
