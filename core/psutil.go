package core

import (
	"fmt"
	"github.com/phuslu/log"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"strconv"
	"time"
)

func GetPcInfo() (info PcInfo) {
	ch := ProcessStatus.IterBuffered()

	var rv []ProcessInfo
	for item := range ch {
		pid, ok := item.Val.(int)
		if !ok {
			continue
		}
		rv = append(rv, ProcessInfo{
			TaskId: item.Key,
			Pid:    int32(pid),
		})
	}
	info.ProStatus = rv
	info.Version = Version
	info.ProNum = len(rv)
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Error().Msgf("获取内存占用失败: %s", err)
	}
	info.MemTotal = m.Total
	info.MemUsed = m.Used
	cpuOri, err := cpu.Percent(3*time.Second, false)
	if err != nil {
		log.Error().Msgf("获取cpu占用失败: %s", err)
	}
	cpu1, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", cpuOri[0]), 64)
	info.TotalPercent = cpu1

	return info
}
