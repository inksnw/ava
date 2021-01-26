package avah

import (
	"ava/core"
	"github.com/phuslu/log"
)

func taskrouter(p core.TaskMsg) {
	cmd := allConfig[p.Worker].Command
	log.Info().Msgf("接收到原始参数: %s  %s  %s", p.Worker, p.Route, p.TaskID)
	dir := allConfig[p.Worker].Dir
	go executor(cmd, p.Params, p.TaskID, dir)
}
