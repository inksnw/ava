package avah

import (
	"ava/core"
	"fmt"
	"github.com/phuslu/log"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"
)

var allConfig map[string]core.LauncherConf
var taskChan = make(chan map[string]core.LauncherConf, 1024)

func sendMsg() {
	for {
		msg := <-taskChan
		err := connIns.WJson(msg)
		if err != nil {
			log.Error().Msgf("更新节点信息失败 %s", err)
			return
		}
	}
}

func listAll(path string) {
	files, err := ioutil.ReadDir(path)
	var allConfig = make(map[string]core.LauncherConf)
	if err != nil {
		panic(fmt.Errorf("遍历目录失败: %s \n", err))
	}
	for _, fi := range files {
		if fi.IsDir() {
			//log.Debug().Msgf("解析%s目录下的配置文件", fi.Name())
			var config core.LauncherConf
			viper.SetConfigFile(filepath.Join(fi.Name(), "launcher1.json"))
			err := viper.ReadInConfig() // 读取配置数据
			if err != nil {
				//log.Debug().Msgf("目录: %s下没有找到%s,跳过", fi.Name(), confname)
				continue
			}
			viper.Unmarshal(&config) // 将配置信息绑定到结构体上
			allConfig[config.Worker] = config
		}
	}
	taskChan <- allConfig
}
