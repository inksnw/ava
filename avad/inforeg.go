package avad

import (
	"ava/core"
	"github.com/phuslu/log"
	"time"
)

var workerMap = make(map[string][]string)
var workerMapR = make(map[string][]string)
var AllInfo = make(map[string]core.PcInfo)

func dealConfig(host string, p map[string]core.LauncherConf) {

	for k, _ := range p {
		workerMapR[host] = append(workerMapR[host], k)
	}
	for k, _ := range p {
		workerMap[k] = append(workerMap[k], host)
	}
	//去重配置
	for k, v := range workerMap {
		workerMap[k] = RemoveRepeatedElement(v)
	}
	for k, v := range workerMapR {
		workerMapR[k] = RemoveRepeatedElement(v)
	}

}

func getNodeInfo(host string, ins *core.WsStruct) {
	//todo map并发安全

	for {
		p := make(map[string]core.LauncherConf)
		_ = ins.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		err := ins.RJson(&p)

		if err != nil {
			log.Error().Msgf("读取节点: %s信息失败,立刻重连 %s,", host, err)
			Reconnect(ins, host)
			return
		}
		if value, ok := p["info"]; ok {
			AllInfo[host] = value.PcInfo
			//log.Info().Msgf("读取节点: %s 状态信息成功", host)
			continue
		}
		dealConfig(host, p)
		//log.Info().Msgf("读取节点: %s注册信息成功 +%v", host,p)
	}

}

func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return newArr
}
