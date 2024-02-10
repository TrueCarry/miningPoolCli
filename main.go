package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"miningPoolCli/config"
	"miningPoolCli/utils/api"
	"miningPoolCli/utils/gpuwrk"
	"miningPoolCli/utils/helpers"
	"miningPoolCli/utils/initp"
	"miningPoolCli/utils/logreport"
	"miningPoolCli/utils/mlog"
	"miningPoolCli/utils/server"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

var gpuGoroutines []gpuwrk.GpuGoroutine
var globalTasks []api.Task

func startTask(i int, task api.Task) {
	// gpuGoroutines[i].startTimestamp = time.Now().Unix()

	if task.Expire < time.Now().Unix() {
		if gpuGoroutines[i].KeepAlive {
			enableTask(i)

		}
		return
	}
	minerArgs := []string{
		// "-vv",
		// "-V",
		// "-B",
		"-g" + strconv.Itoa(gpuGoroutines[i].GpuData.GpuId),
		"-p" + strconv.Itoa(gpuGoroutines[i].GpuData.PlatformId),
		"-F" + strconv.Itoa(config.StaticBeforeMinerSettings.BoostFactor),
		"-t" + strconv.Itoa(config.StaticBeforeMinerSettings.TimeoutT),
		// "-e" + strconv.FormatInt(task.Expire, 10),
		config.StaticBeforeMinerSettings.PoolAddress,
		helpers.ConvertHexData(task.Seed),
		helpers.ConvertHexData(task.Complexity),
		config.StaticBeforeMinerSettings.Iterations,
		// task.Giver,
		// pathToBoc,
	}
	cmd := exec.Command(gpuGoroutines[i].GpuData.StartPath, minerArgs...)

	gpuGoroutines[i].ProcStderr.Reset()
	cmd.Stderr = &gpuGoroutines[i].ProcStderr

	unblockFunc := make(chan struct{}, 1)

	var killedByNotActual bool
	var done bool

	if err := cmd.Start(); err != nil {
		mlog.LogFatal("failed to start miner cmd; err: " + err.Error() + "; args: " + strings.Join(cmd.Args, " "))
	}

	gpuGoroutines[i].PPid = cmd.Process.Pid
	gpuGoroutines[i].KeepAlive = true

	go func() {
		cmd.Wait()
		done = true

		out := gpuGoroutines[i].ProcStderr.String()
		lines := strings.Split(out, "\n")
		if len(lines) > 3 {
			if strings.Contains(lines[len(lines)-3], "FOUND!") {
				if !killedByNotActual {
					go func() {
						lineWithProof := lines[len(lines)-2]
						if len(lineWithProof) < 246 {
							mlog.LogError("Decoding proof len: " + strconv.Itoa(len(lineWithProof)))
							return
						}
						hexData, err := hex.DecodeString(lineWithProof[:246])
						if err != nil {
							mlog.LogError("Decoding proof error: " + strconv.Itoa(len(lineWithProof)) + " " + err.Error())
							return
						}

						body := cell.BeginCell().MustStoreBinarySnake(hexData[2:]).EndCell()

						giverAddress, _ := address.ParseAddr(task.Giver)
						addrData := giverAddress.Data()
						extCell := cell.BeginCell().
							MustStoreUInt(0x44, 7).
							MustStoreUInt(uint64(giverAddress.Workchain()), 8). // giver workchain
							MustStoreBinarySnake(addrData).
							MustStoreUInt(1, 6). // amount grams
							MustStoreRef(body).
							EndCell()

						bocServerResp, err := api.SendHexBocToServer(hex.EncodeToString(extCell.ToBOC()), task.Seed, strconv.Itoa(task.Id))
						if err == nil {
							if bocServerResp.Data == "Found" && bocServerResp.Status == "ok" {
								logreport.ShareFound(gpuGoroutines[i].GpuData.Model, gpuGoroutines[i].GpuData.GpuId, task.Id)
							} else {
								logreport.ShareServerError(task, bocServerResp, gpuGoroutines[i].GpuData.GpuId)
							}
						}
					}()
				}
			}
		} else {
			mlog.LogInfo(fmt.Sprintf(
				"Working, no shares found. Everythging is OK",
			))
		}

		if gpuGoroutines[i].KeepAlive {
			enableTask(i)
		}

		unblockFunc <- struct{}{}
	}()

	go func() {
		for !done {
			// Task no longer in list, kill
			existing := checkTaskAlreadyFound(task.Id)
			if existing == nil {
				// log.Println("Task expired")
				killedByNotActual = true
				if err := cmd.Process.Kill(); err != nil {
					mlog.LogError(err.Error())
				}
				break
				// Task expired, kill
			} else if existing.Expire < time.Now().Unix() {
				// log.Println("Task expired", task.Expire, time.Now().Unix(), time.Now().Unix()-task.Expire)
				killedByNotActual = true
				if err := cmd.Process.Kill(); err != nil {
					mlog.LogError(err.Error())
				}
				break
			}
			time.Sleep(64 * time.Millisecond)
		}
	}()

	<-unblockFunc
}

func checkTaskAlreadyFound(checkId int) *api.Task {
	for _, task := range globalTasks {
		if task.Id == checkId {
			return &task
		}
	}
	return nil
}

func syncTasks(firstSync *chan struct{}) {
	var firstSyncIsOk bool
	for {
		if v, err := api.GetTasks(); err == nil && len(v.Tasks) > 0 {
			globalTasks = v.Tasks
			if !firstSyncIsOk {
				*firstSync <- struct{}{}
				firstSyncIsOk = true
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func enableTask(gpuGoIndex int) {
	if tLen := len(globalTasks); tLen > 0 {
		go startTask(gpuGoIndex, globalTasks[rand.Intn(len(globalTasks))])
	} else {
		mlog.LogError("can't start task, because the len of globalTasks <= 0")
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	gpus := initp.InitProgram()

	gpuGoroutines = make([]gpuwrk.GpuGoroutine, len(gpus))

	firstSync := make(chan struct{})
	go syncTasks(&firstSync)
	<-firstSync

	for gpuGoIndex := range gpuGoroutines {
		gpuGoroutines[gpuGoIndex].GpuData = gpus[gpuGoIndex]
		enableTask(gpuGoIndex)
	}

	if !config.NetSrv.RunThis && config.NetSrv.HandleKill {
		mlog.LogInfo("Unable to apply -handle-kill because flag -serve-stat is not specified")
	} else if config.NetSrv.RunThis {
		go server.Entrypoint(&gpuGoroutines)
	}

	for {
		time.Sleep(1 * time.Second)
		gpuwrk.CalcHashrate(&gpuGoroutines)
	}
}
