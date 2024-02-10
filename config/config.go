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

package config

import (
	"regexp"
	"time"
)

type colors struct {
	ColorRed, ColorGreen,
	ColorYellow, ColorBlue,
	ColorReset string
}

type minerGetter struct {
	MinerDirectory string
	UbuntuSettings struct {
		ReleaseURL, FileName, ExecutableName, ExecutableNameCuda string
	}
	WinSettings struct {
		ReleaseURL, FileName, ExecutableName, ExecutableNameCuda string
	}
	MacSettings struct {
		ReleaseURL, FileName, ExecutableName, ExecutableNameCuda string
	}
	ExecNamePref       string // "./" in linux; "" in win
	CurrExecNameOpenCL string // current ExecutableName (depends on OS)
	CurrExecNameCuda   string // current ExecutableName (depends on OS)
	// StartPath          string // depends on OS
}

type os struct {
	OperatingSystem, Architecture string
}

type serverSettings struct {
	MiningPoolServerURL, AuthKey string
}

type staticBeforeMinerSettings struct {
	BoostFactor int
	TimeoutT    int
	Iterations  string
	PoolAddress string
}

type osType struct {
	Linux, Win, Macos string
}

type minerRegexKit struct {
	FindGPUPat, ReplaceStartGPU, ReplaceEndGPU,
	FindIntIds, FindHashRate, FindDecimal *regexp.Regexp
}

// miner server config
type netServer struct {
	Host         string
	HostFileName string
	RunThis      bool
	HandleKill   bool
}

var Colors colors
var MinerGetter minerGetter
var OS os
var ServerSettings serverSettings
var StaticBeforeMinerSettings staticBeforeMinerSettings
var OSType osType
var MRgxKit minerRegexKit

var UpdateStatsFile bool
var StartProgramTimestamp int64

var NetSrv netServer

func Configure() {
	// -------- minerRegexKit
	MRgxKit = minerRegexKit{
		FindGPUPat:      regexp.MustCompile(`(\[ [^\]]+ \])`),
		ReplaceStartGPU: regexp.MustCompile(`^\[ (OpenCL: platform #[0-9]+ device #[0-9]+|GPU #[0-9]+:)`),
		ReplaceEndGPU:   regexp.MustCompile(`\](.*)`),
		FindIntIds:      regexp.MustCompile(`#\d[\d,]*`),
		FindHashRate:    regexp.MustCompile(`instant speed: (\d+\.?\d*) Mhash\/s`),
		FindDecimal:     regexp.MustCompile(`(\d+\.?\d*)`),
	}
	// --------

	StartProgramTimestamp = time.Now().Unix()

	Colors = colors{
		"\u001b[31m", "\u001b[32m",
		"\u001b[33m", "\u001b[34m",
		"\u001b[0m",
	}

	// -------- OS Types
	OSType = osType{
		Linux: "linux",
		Win:   "windows",
		Macos: "darwin",
	}
	// --------

	MinerGetter.MinerDirectory = "miner_blob"

	// -------- set Release for Ubuntu
	MinerGetter.UbuntuSettings.FileName = "minertools-cuda-ubuntu-18.04-x86-64.tar.gz"
	MinerGetter.UbuntuSettings.ReleaseURL = "https://ton.ninja/miners/" +
		MinerGetter.UbuntuSettings.FileName
	MinerGetter.UbuntuSettings.ExecutableName = "pow-miner-opencl"
	MinerGetter.UbuntuSettings.ExecutableNameCuda = "pow-miner-cuda"
	// --------

	// -------- set Release for Win
	MinerGetter.WinSettings.FileName = "minertools-cuda-windows-x86-64.zip"
	MinerGetter.WinSettings.ReleaseURL = "https://ton.ninja/miners/" +
		MinerGetter.WinSettings.FileName
	MinerGetter.WinSettings.ExecutableName = "pow-miner-opencl.exe"
	MinerGetter.WinSettings.ExecutableNameCuda = "pow-miner-cuda.exe"
	// --------

	// -------- set Release for Mac
	MinerGetter.MacSettings.FileName = ""
	MinerGetter.MacSettings.ReleaseURL = "https://raw.githubusercontent.com/TrueCarry/JettonGramGpuMiner/main/pow-miner-opencl-macos"
	MinerGetter.MacSettings.ExecutableName = "pow-miner-opencl-macos"
	// --------

	// -------- StaticBeforeMinerSettings
	StaticBeforeMinerSettings = staticBeforeMinerSettings{
		BoostFactor: 512,
		Iterations:  "9223372036854775807",
		TimeoutT:    15,
	}
	// --------

	// -------- Net server
	NetSrv = netServer{
		Host:         "127.0.0.1",
		HostFileName: "serveraddr.txt",
	}
	// --------

	// -------- configure texts
	configureTexts()
	// --------
}
