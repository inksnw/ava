package avah

import (
	"ava/core"
	"fmt"
	"github.com/phuslu/log"
)

func taskrouter(p core.TaskMsg) {
	cmd := allConfig[p.Worker].Command
	var route string
	if p.Route != "" {
		route = fmt.Sprintf("定点投送 %s", p.Route)
	}

	log.Info().Msgf("接收到原始参数: %s  %s  %s %s", cmd, p.Worker, p.TaskID, route)
	dir := allConfig[p.Worker].Dir
	go executor(cmd, p.Params, p.TaskID, dir)
}
