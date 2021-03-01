package core

import (
	cmap "github.com/orcaman/concurrent-map"
	"time"
)

const WsPort = "4560"
const TcpPort = "4561"
const SocksPort = "4562"
const Web = "4000"

type LauncherConf struct {
	Worker  string `json:"worker"`
	Command string `json:"command"`
	Dir     string `json:"dir"`
	PcInfo
}

var ProcessStatus = cmap.New()

type ProcessInfo struct {
	TaskId string
	Pid    int32
}

type PcInfo struct {
	Version      string `json:"Version"`
	ProNum       int
	MemTotal     uint64
	MemUsed      uint64
	TotalPercent float64
	ProStatus    []ProcessInfo
}

type TaskMsg struct {
	Route  string `json:"route"`
	Worker string `json:"worker"`
	TaskID string `json:"task_id"`
	Params string `json:"params"`
	Rand   bool   `json:"rand"`
}

const PongWait = 20 * time.Second
const UpdateWait = 5 * time.Second

//每台主机最多运行业务进程数
var PerMachineProcess int

//白名单
var Sites []string

var Version = "1.21"
