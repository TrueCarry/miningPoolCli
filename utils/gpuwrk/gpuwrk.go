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

package gpuwrk

import (
	"bytes"
	"errors"
	"miningPoolCli/config"
	"miningPoolCli/utils/mlog"
	"os/exec"
	"strconv"
	"strings"
)

type GPUstruct struct {
	GpuId      int    `json:"device_id"`
	Model      string `json:"device_name"`
	PlatformId int    `json:"platform_id"`
	StartPath  string `json:"start_path"`
}

type GpuGoroutine struct {
	GpuData GPUstruct
	// startTimestamp int64
	CurrentHashrate int

	ProcStderr bytes.Buffer
	PPid       int
	KeepAlive  bool
}

func LogGpuList(gpus []GPUstruct) {
	var gpuNames []string

	for i := 0; i < len(gpus); i++ {
		gpuNames = append(gpuNames, gpus[i].Model)
	}

	dict := make(map[string]int)
	for _, num := range gpuNames {
		dict[num] = dict[num] + 1
	}

	mlog.LogInfo("Found GPUs:")
	for model, count := range dict {
		mlog.LogInfo("x" + strconv.Itoa(count) + " " + model + "\n")
	}
}

func searchGpusWithRegexCuda(execStr string) ([]GPUstruct, error) {
	cmd := exec.Command(execStr)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		mlog.LogFatalStackError(err)
	}
	cmd.Wait()

	matches := config.MRgxKit.FindGPUPat.FindAllString(stderr.String(), -1)

	var gpusArray []GPUstruct

	for _, v := range matches {
		gpuModel := strings.TrimSpace(
			config.MRgxKit.ReplaceEndGPU.ReplaceAllString(config.MRgxKit.ReplaceStartGPU.ReplaceAllString(v, ""), ""),
		)

		// skip for integrated intel GPU
		if strings.Contains(strings.ToLower(gpuModel), "intel") {
			continue
		}

		panddId := config.MRgxKit.FindIntIds.FindAllString(v, -1)
		if len(panddId) < 1 {
			mlog.LogInfo("warn: can't search GPU - len(panddId) < 2; panddId: " + strings.Join(panddId, ", ") + "; v: " + v)
			continue
		}

		// platformId, err := strconv.Atoi(strings.Replace(panddId[0], "#", "", -1))
		// if err != nil {
		// 	return gpusArray, errors.New("can't get platformId: " + err.Error())
		// }
		deviceId, err := strconv.Atoi(strings.Replace(panddId[0], "#", "", -1))
		if err != nil {
			return gpusArray, errors.New("can't get deviceId: " + err.Error())
		}

		gpusArray = append(
			gpusArray,
			GPUstruct{
				GpuId:      deviceId,
				Model:      gpuModel,
				PlatformId: 0,
			},
		)
	}

	// if len(gpusArray) < 1 {
	// 	return gpusArray, errors.New("no any GPUs found")
	// }

	return gpusArray, nil
}

func SearchGpusCuda(execStr string) []GPUstruct {
	nvidiaGpus, err := searchGpusWithRegexCuda(execStr)
	if err != nil {
		mlog.LogFatal(err.Error())
	}
	return nvidiaGpus
}

func searchGpusWithRegexOpenCL(execStr string) ([]GPUstruct, error) {
	cmd := exec.Command(execStr)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		mlog.LogFatalStackError(err)
	}
	cmd.Wait()

	matches := config.MRgxKit.FindGPUPat.FindAllString(stderr.String(), -1)

	var gpusArray []GPUstruct

	for _, v := range matches {
		gpuModel := strings.TrimSpace(
			config.MRgxKit.ReplaceEndGPU.ReplaceAllString(config.MRgxKit.ReplaceStartGPU.ReplaceAllString(v, ""), ""),
		)

		// skip for integrated intel GPU
		if strings.Contains(strings.ToLower(gpuModel), "intel") {
			continue
		}

		panddId := config.MRgxKit.FindIntIds.FindAllString(v, -1)
		if len(panddId) < 2 {
			mlog.LogInfo("warn: can't search GPU - len(panddId) < 2; panddId: " + strings.Join(panddId, ", ") + "; v: " + v)
			continue
		}

		platformId, err := strconv.Atoi(strings.Replace(panddId[0], "#", "", -1))
		if err != nil {
			return gpusArray, errors.New("can't get platformId: " + err.Error())
		}
		deviceId, err := strconv.Atoi(strings.Replace(panddId[1], "#", "", -1))
		if err != nil {
			return gpusArray, errors.New("can't get deviceId: " + err.Error())
		}

		gpusArray = append(
			gpusArray,
			GPUstruct{
				GpuId:      deviceId,
				Model:      gpuModel,
				PlatformId: platformId,
			},
		)
	}

	// if len(gpusArray) < 1 {
	// 	return gpusArray, errors.New("no any GPUs found")
	// }

	return gpusArray, nil
}

func SearchGpusOpenCL(execStr string) []GPUstruct {
	nvidiaGpus, err := searchGpusWithRegexOpenCL(execStr)
	if err != nil {
		mlog.LogFatal(err.Error())
	}
	return nvidiaGpus
}
