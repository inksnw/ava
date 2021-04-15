package avah

import (
	"ava/core"
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/phuslu/log"
)

func updateProcess() {
	tmp := make(map[string]core.LauncherConf)
	tmp["info"] = core.LauncherConf{
		PcInfo: core.GetPcInfo(),
	}
	msgChan <- tmp
}

func executor(command, arg, taskid, dir string) {

	runLeft := time.Duration(30) * time.Minute
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(runLeft))
	filename := tmpConfig(dir, arg, taskid)

	script := strings.Split(command, " ")
	script = append(script, "placeholder", filename)

	// 开启脚本子进程
	log.Debug().Msgf("工作目录 %s", dir)
	cmd := exec.CommandContext(ctx, script[0], script[1:]...)
	cmd.Dir = dir

	// 开启脚本子进程 stdout 和 stderr 并合并
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		log.Error().Msgf("开启日志管道失败")
	}

	// 启动 脚本子进程
	err = cmd.Start()
	if err != nil {
		log.Error().Msgf("程序启动失败,任务id: %s,%s,%s", taskid, script, err)
		return
	}

	// 开启 stdout buffer
	// 开启 log文件 io
	scanner := bufio.NewScanner(stdout)
	logfile := filepath.Join(dir, taskid+".log")
	fileio, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Error().Msgf("开启日志文件失败, Err: %v", err)
		return
	}

	// 读取 脚本 进程输出到log文件
	go func() {
		for scanner.Scan() {
			bs := scanner.Bytes()
			_, err := fileio.Write(bs)
			if err != nil {
				log.Error().Msgf("脚本stdout写入日志文件失败, Err: %v", err)
			}
		}
	}()

	log.Info().Msgf("程序启动成功,任务id: %s 进程id: %d", taskid, cmd.Process.Pid)
	tmp := core.PidStruct{
		Pid:        cmd.Process.Pid,
		CreateTime: time.Now().Unix(),
	}

	core.ProcessStatus.Set(taskid, tmp)
	updateProcess()

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Error().Msgf("进程 %s异常退出 %s", logfile, err)
			updateProcess()
			e := fmt.Sprintf("进程异常退出 %s\n", err)
			_, err = fileio.Write([]byte(e))
			if err != nil {
				log.Error().Msgf("进程 %s异常退出,写入日志失败 %s", logfile, err)
			}
		}
		log.Debug().Msgf("任务: %s 执行完成,退出", command)
		updateProcess()
		core.ProcessStatus.Remove(taskid)
		fileio.Close()
	}()
}

func tmpConfig(dir, arg, taskid string) (filename string) {
	tmpFile, err := ioutil.TempFile(dir, taskid)
	if err != nil {
		log.Error().Msgf("创建文件型参数失败: %s", err)
		return ""
	}
	_, err = tmpFile.Write([]byte(arg))
	if err != nil {
		log.Error().Msgf("向参数文件写入内容失败 %s", err)
		return ""
	}
	return tmpFile.Name()
}
