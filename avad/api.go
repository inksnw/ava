package avad

import (
	"ava/core"
	"encoding/json"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/phuslu/log"
	"io/ioutil"
	"net/http"
)

func webWsStatus(w http.ResponseWriter, r *http.Request) {
	rv := map[string]cmap.ConcurrentMap{}
	rv["ws"] = wsStatus
	rv["tcp"] = tcpStatus
	err := json.NewEncoder(w).Encode(rv)
	if err != nil {
		log.Error().Msgf("返回前端接口失败 %s", err)
	}

}

func webWorkerMapR(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(workerMapR)
	if err != nil {
		log.Error().Msgf("返回前端接口失败 %s", err)
	}

}

func info(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(AllInfo)
	if err != nil {
		log.Error().Msgf("返回前端接口失败 %s", err)
	}

}

type webInfo struct {
	Host     string   `json:"Host"`
	Business []string `json:"Business"`
	Status   bool     `json:"Status"`
	core.PcInfo
}

func getAllInfo(w http.ResponseWriter, r *http.Request) {
	var rv []webInfo

	ch := wsStatus.IterBuffered()
	for item := range ch {
		host := item.Key
		tmp, _ := wsStatus.Get(host)
		sta := tmp.(bool)
		infoOne := webInfo{
			Host:     host,
			Business: workerMapR[host],
			Status:   sta,
			PcInfo:   AllInfo[host],
		}
		rv = append(rv, infoOne)
	}
	err := json.NewEncoder(w).Encode(rv)
	if err != nil {
		log.Error().Msgf("返回前端接口失败 %s", err)
	}
}

func getProxyInfo(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	ip := r.FormValue("ip")
	check := r.FormValue("check")
	var url string
	if check == "1" {
		url = "http://" + ip + ":" + "6543" + "/" + "check"
		log.Debug().Msgf("转发测试 %s 的代理状态", url)
	} else {
		url = "http://" + ip + ":" + "6543"
		log.Debug().Msgf("转发查看 %s 的代理信息", url)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("请求%s失败", url)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	_, err = w.Write(body)
	if err != nil {
		log.Error().Msgf("返回前端接口失败 %s", err)
	}
	defer resp.Body.Close()

}
