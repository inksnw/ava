package avad

import (
	"ava/core"
	"fmt"
	"github.com/phuslu/log"
	"math/rand"
	"sort"
	"time"
)

func fixed(p core.TaskMsg) (rv *result) {
	log.Debug().Msgf("接到 %s的定点任务", p.Route)
	_, err := netAvailable(p.Route)
	if err != nil {
		return &result{400, err.Error(), p.Route}
	}
	host := workerAvailable(p.Route, p.Worker)
	if host != "" {
		code, msg := send(p.Route, p)

		if code == 400 {
			log.Debug().Msgf("节点%s状态更新可能延迟,再次重试投送", host)
			fixed(p)
		}

		return &result{code, msg, p.Route}
	}
	msg := fmt.Sprintf("未在主机: %s上找到业务%s,请检查参数", p.Route, p.Worker)
	return &result{400, msg, p.Route}
}

func balance(p core.TaskMsg) (rv *result) {
	log.Debug().Msgf("自动解析: %s任务的运行节点", p.Worker)
	hosts, ok := workerMap[p.Worker]
	if !ok {
		msg := fmt.Sprintf("投送的业务: %s未找到", p.Worker)
		return &result{400, msg, ""}
	}
	var host string
	if p.Rand {
		host = randOne(hosts, p)
	} else {
		host = balanceOne(hosts, p)
	}
	if host == "" {
		msg := fmt.Sprintf("任务: %s在节点 %s都有部署,全节点不可达", p.Route, hosts)
		return &result{400, msg, ""}
	}
	code, msg := send(host, p)

	if host != "" && code == 400 {
		log.Debug().Msgf("节点%s状态更新可能延迟,再次重试投送", host)
		balance(p)
	}

	msg = fmt.Sprintf("任务%s在%s都有部署, %s", p.Worker, hosts, msg)
	return &result{code, msg, host}

}

type machine struct {
	ip     string
	proNum int
}

func balanceOne(hosts []string, p core.TaskMsg) (host string) {
	var allMachine []machine
	for k, v := range AllInfo {
		//仅在有这个业务的主机里寻找
		if core.StringInSlice(k, hosts) {
			allMachine = append(allMachine, machine{k, v.ProNum})
		}
	}
	sort.Slice(allMachine, func(i, j int) bool {
		//return allMachine[i].proNum > allMachine[j].proNum  // 降序
		return allMachine[i].proNum < allMachine[j].proNum // 升序
	})

	var workderMachine []string

	for _, v := range allMachine {
		host = v.ip
		workderMachine = append(workderMachine, host)
		if _, err := netAvailable(host); err != nil {
			log.Info().Msgf("任务[%s:%s]在节点 %s都有部署,任务数最低的节点: %s不可用,更换下一个", p.Worker, p.TaskID, workderMachine, host)
			continue
		}
		log.Info().Msgf("任务: [%s:%s]在节点 %s都有部署,投送到任务数最低的节点: %s", p.Worker, p.TaskID, workderMachine, host)
		return host
	}
	return ""
}

func randOne(hosts []string, p core.TaskMsg) (host string) {
	tmp := make([]string, len(hosts), len(hosts))
	copy(tmp, hosts)
	for range tmp {
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s) // initialize local pseudorandom generator
		index := r.Intn(len(tmp))
		host = tmp[index]
		if _, err := netAvailable(host); err != nil {
			log.Info().Msgf("任务[%s:%s]在节点 %s都有部署,随机节点: %s不可用,更换下一个", p.Worker, p.TaskID, hosts, host)
			tmp = append(tmp[:index], tmp[index+1:]...)
			continue
		}
		log.Info().Msgf("任务:[%s:%s]在节点 %s都有部署,随机投送到: %s执行", p.Worker, p.TaskID, hosts, host)
		return host
	}
	return ""
}
