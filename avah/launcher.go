package avah

import (
	"ava/core"
	"context"
	"fmt"
	"github.com/phuslu/log"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(runLeft))
	filename := tmpConfig(dir, arg, taskid)

	script := strings.Split(command, " ")
	script = append(script, "placeholder", filename)

	log.Debug().Msgf("工作目录 %s", dir)
	cmd := exec.CommandContext(ctx, script[0], script[1:]...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		log.Error().Msgf("开启日志管道失败")
	}

	if err = cmd.Start(); err != nil {
		log.Error().Msgf("程序启动失败,任务id: %s,%s,%s", taskid, script, err)
		//if err := os.Remove(filename); err != nil {
		//	log.Error().Msgf("程序启动失败,临时参数文件删除失败 %s", err)
		//}
		cancel()
		return
	}

	logfile := filepath.Join(dir, taskid+".log")
	dstLog := createFile(logfile)

	// 从管道中实时获取输出并打印到终端
	go asyncLog(ctx, stdout, dstLog)

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
			_, err = dstLog.Write([]byte(e))
			if err != nil {
				log.Error().Msgf("进程 %s异常退出,写入日志失败 %s", logfile, err)
			}
		}
		cancel()
		//if err := os.Remove(filename); err != nil {
		//	log.Error().Msgf("任务执行完成,临时参数文件删除失败 %s", err)
		//}
		log.Debug().Msgf("任务: %s 执行完成,退出", command)
		updateProcess()
		core.ProcessStatus.Remove(taskid)
		dstLog.Close()
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

func createFile(filename string) (file *os.File) {
	var err error
	if Exists(filename) {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Error().Msgf("文件%s打开失败%s", filename, err)
			return nil
		}
	} else {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Error().Msgf("文件%s创建失败%s", filename, err)
			return nil
		}
	}
	log.Debug().Msgf("日志文件%s创建成功", filename)
	return file
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func asyncLog(ctx context.Context, stdout io.ReadCloser, dstLog *os.File) {
	buf := make([]byte, 1024, 1024)
	//defer dstLog.Close()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("进程结束日志协程退出")
			return
		default:
			strNum, err := stdout.Read(buf)
			if strNum > 0 {
				outputByte := buf[:strNum]
				_, err = dstLog.Write(outputByte)
				if err != nil {
					log.Debug().Msgf("%s 日志文件写入失败 %s", dstLog.Name(), err)
					return
				}
			}
			if err != nil {
				//读到结尾
				if err == io.EOF || strings.Contains(err.Error(), "file already closed") {
					err = nil
				}
			}
		}
	}
}
