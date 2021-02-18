package avad

import (
	"ava/core"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/phuslu/log"
	"net/http"
)

type result struct {
	Code  int
	Msg   string
	Route string
}

func resourceAvailable() (totalTasks, currentTasks int, available bool) {
	totalTasks = core.PerMachineProcess * len(AllInfo)
	for _, v := range AllInfo {
		currentTasks = currentTasks + v.ProNum
	}
	if core.PerMachineProcess == 0 {
		//未配置单主机限制
		log.Debug().Msgf("未配置单主机任务数限制,全部放行")
		available = true
		return
	}

	if currentTasks >= totalTasks {
		available = false
		return
	}
	return totalTasks, currentTasks, true
}

func taskRouter(w http.ResponseWriter, r *http.Request) {
	var p core.TaskMsg
	var rv = &result{}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	totalTasks, currentTasks, available := resourceAvailable()
	if !available {
		msg := fmt.Sprintf("单节点 %d个任务 * 在线节点数 %d 共能执行%d个任务,已运行%d个,无法承载,请稍后再试", core.PerMachineProcess, len(AllInfo), totalTasks, currentTasks)
		log.Debug().Msgf(msg)
		rv = &result{503, msg, p.Route}
		json.NewEncoder(w).Encode(rv)
		return
	}

	if p.Route != "" {
		rv = fixed(p)

	} else {
		rv = balance(p)
	}
	json.NewEncoder(w).Encode(rv)

}

func workerAvailable(host, workerdst string) (hostdst string) {
	if workers, ok := workerMapR[host]; ok {
		for _, worker := range workers {
			if worker == workerdst {
				return host
			}
		}
		return ""
	}
	return ""
}

func netAvailable(host string) (conn *websocket.Conn, err error) {

	ins, ok := ConnStatus.Get(host)
	if !ok {
		return nil, fmt.Errorf("未找到节点: %s,请检查输入", host)
	}
	instance := ins.(*ConnStruct)
	if !instance.status {
		return nil, fmt.Errorf("节点: %s,网络中断", host)
	}
	return instance.conn, nil

}

func send(host string, p core.TaskMsg) (code int, msg string) {
	conn, err := netAvailable(host)
	if err != nil {
		return 400, fmt.Sprintf("%s", err)
	}

	ins, _ := ConnStatus.Get(host)
	instance := ins.(*ConnStruct)

	if conn != nil {
		log.Info().Msgf("发送前原始参数: %s  %s  %s", p.Worker, p.Route, p.TaskID)
		err = conn.WriteJSON(p)
		if err != nil {
			log.Error().Msgf("投送失败,节点: %s可能已不在线", host)
			instance.status = false
			err = instance.conn.Close()
			if err != nil {
				log.Error().Msgf("关闭连接失败: %s", err)
			}
			return 400, fmt.Sprintf("投送失败,节点: %s可能已不在线 %s", host, err)
		}
	}
	return 200, fmt.Sprintf("投送到: %s成功", host)
}
