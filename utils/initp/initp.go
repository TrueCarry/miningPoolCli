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

package initp

import (
	"flag"
	"fmt"
	"log"
	"miningPoolCli/config"
	"miningPoolCli/utils/api"
	"miningPoolCli/utils/getminer"
	"miningPoolCli/utils/gpuwrk"
	"miningPoolCli/utils/mlog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func InitProgram() []gpuwrk.GPUstruct {
	config.Configure()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, config.Texts.GlobalHelpText)
	}

	flag.StringVar(&config.ServerSettings.AuthKey, "pool-id", "", "")
	flag.StringVar(&config.ServerSettings.MiningPoolServerURL, "url", "https://ninja.tonlens.com", "")
	flag.BoolVar(&config.UpdateStatsFile, "stats", false, "") // for Hive OS

	flag.BoolVar(&config.NetSrv.RunThis, "serve-stat", false, "")     // run http server with miner stat
	flag.BoolVar(&config.NetSrv.HandleKill, "handle-kill", false, "") // handle /kill (os.Exit by http)

	flag.Parse()
	config.OS.OperatingSystem, config.OS.Architecture = runtime.GOOS, runtime.GOARCH

	switch "" {
	case config.ServerSettings.AuthKey:
		mlog.LogFatal("Flag -pool-id is required; for help run with -h flag")
	}

	mlog.LogText(config.Texts.Logo)
	mlog.LogText(config.Texts.WelcomeAdditionalMsg)

	supportedOS := false
	if config.OS.OperatingSystem == config.OSType.Win && config.OS.Architecture == "amd64" {
		supportedOS = true
	}
	if config.OS.OperatingSystem == config.OSType.Linux && config.OS.Architecture == "amd64" {
		supportedOS = true
	}
	if config.OS.OperatingSystem == config.OSType.Macos && config.OS.Architecture == "arm64" {
		supportedOS = true
	}

	if supportedOS {
		mlog.LogOk("Supported OS detected: " + config.OS.OperatingSystem + "/" + config.OS.Architecture)
	} else {
		mlog.LogFatal("Unsupported OS detected: " + config.OS.OperatingSystem + "/" + config.OS.Architecture)
	}

	mlog.LogInfo("Using mining pool API url: " + config.ServerSettings.MiningPoolServerURL)

	switch config.OS.OperatingSystem {
	case config.OSType.Linux:
		config.MinerGetter.CurrExecNameOpenCL = config.MinerGetter.UbuntuSettings.ExecutableName
		config.MinerGetter.CurrExecNameCuda = config.MinerGetter.UbuntuSettings.ExecutableNameCuda
		config.MinerGetter.ExecNamePref = "./"
	case config.OSType.Win:
		config.MinerGetter.CurrExecNameOpenCL = config.MinerGetter.WinSettings.ExecutableName
		config.MinerGetter.CurrExecNameCuda = config.MinerGetter.WinSettings.ExecutableNameCuda
		config.MinerGetter.ExecNamePref = ""
	case config.OSType.Macos:
		config.MinerGetter.CurrExecNameOpenCL = config.MinerGetter.MacSettings.ExecutableName
		config.MinerGetter.CurrExecNameCuda = config.MinerGetter.MacSettings.ExecutableNameCuda
		config.MinerGetter.ExecNamePref = "./"
	}

	for {
		if ok := api.Auth(); ok {
			break
		}
		mlog.LogInfo("Auth attempt no successfull, waiting 5 seconds")
		time.Sleep(time.Second * 5)
	}

	getminer.GetMiner()

	var gpusArray []gpuwrk.GPUstruct
	if config.MinerGetter.CurrExecNameCuda != "" {
		if gpusArray = gpuwrk.SearchGpusCuda(filepath.Join(
			config.MinerGetter.ExecNamePref,
			config.MinerGetter.MinerDirectory,
			config.MinerGetter.CurrExecNameCuda,
		)); len(gpusArray) > 0 {
			config.MinerGetter.StartPath = filepath.Join(
				config.MinerGetter.ExecNamePref,
				config.MinerGetter.MinerDirectory,
				config.MinerGetter.CurrExecNameCuda,
			)
		}
	}

	if len(gpusArray) < 1 && config.MinerGetter.CurrExecNameOpenCL != "" {
		if gpusArray = gpuwrk.SearchGpusOpenCL(filepath.Join(
			config.MinerGetter.ExecNamePref,
			config.MinerGetter.MinerDirectory,
			config.MinerGetter.CurrExecNameOpenCL,
		)); len(gpusArray) > 0 {
			config.MinerGetter.StartPath = filepath.Join(
				config.MinerGetter.ExecNamePref,
				config.MinerGetter.MinerDirectory,
				config.MinerGetter.CurrExecNameOpenCL,
			)
		}
	}
	// else if gpusArray =

	if len(gpusArray) < 1 {
		log.Fatal("Gpus Not Found")
		// return gpusArray, errors.New("no any GPUs found")
	}

	mlog.LogPass()
	gpuwrk.LogGpuList(gpusArray)
	mlog.LogInfo("Launching the mining processes...")

	return gpusArray
}
