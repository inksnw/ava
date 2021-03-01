package avah

import (
	"ava/core"
	"fmt"
	"github.com/phuslu/log"
	"github.com/spf13/viper"
	"io/ioutil"
	"path/filepath"
)

var allConfig = make(map[string]core.LauncherConf)
var msgChan = make(chan map[string]core.LauncherConf, 1024)

func sendMsg() {
	for {
		msg := <-msgChan
		err := connIns.WJson(msg)
		if err != nil {
			log.Error().Msgf("回传信息失败 %s", err)
			return
		}
		//log.Info().Msgf("回传信息成功 %v", msg)
	}
}

func loadConfig(path string) {
	files, err := ioutil.ReadDir(path)

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
			err = viper.Unmarshal(&config)
			if err != nil {
				log.Error().Msgf("解析launcher1失败 %s", err)
			}
			allConfig[config.Worker] = config
		}
	}
	msgChan <- allConfig
	log.Error().Msgf("解析配置文件完成")
}
