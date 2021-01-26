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
)

func executor(command, arg, taskid, dir string) {
	//go reaper.Reap()

	ctx, cancel := context.WithCancel(context.Background())
	filename := tmpConfig(dir, arg)

	script := strings.Split(command, " ")
	log.Info().Msgf("启动器接到命令: %s %s %s %s", script[0], script[1], "placeholder", filename)
	log.Debug().Msgf("工作目录 %s", dir)
	cmd := exec.CommandContext(ctx, script[0], script[1], "placeholder", filename)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		log.Error().Msgf("开启日志管道失败")
	}

	if err = cmd.Start(); err != nil {
		log.Error().Msgf("程序%s %s启动失败,任务id: %s,%s", script[0], script[1], taskid, err)
		if err := os.Remove(filename); err != nil {
			log.Error().Msgf("程序启动失败,临时参数文件删除失败 %s", err)
		}
		cancel()
		return
	}

	logfile := filepath.Join(dir, taskid+".log")
	dstlog := createFile(logfile)

	// 从管道中实时获取输出并打印到终端
	go asyncLog(ctx, stdout, dstlog)

	log.Info().Msgf("程序%s %s成功启动,任务id: %s 进程id: %d", script[0], script[1], taskid, cmd.Process.Pid)
	core.ProcessStatus.Set(taskid, cmd.Process.Pid)

	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Error().Msgf("进程 %s异常退出 %s", logfile, err)
			e := fmt.Sprintf("进程异常退出 %s\n", err)
			dstlog.Write([]byte(e))
		}
		cancel()
		if err := os.Remove(filename); err != nil {
			log.Error().Msgf("任务执行完成,临时参数文件删除失败 %s", err)
		}
		log.Debug().Msgf("任务: %s 执行完成,退出", command)
		core.ProcessStatus.Remove(taskid)
	}()

}

func tmpConfig(dir, arg string) (filename string) {
	tmpFile, err := ioutil.TempFile(dir, "arg-")
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

func asyncLog(ctx context.Context, stdout io.ReadCloser, dstlog *os.File) {
	buf := make([]byte, 1024, 1024)
	defer dstlog.Close()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("进程结束日志协程退出")
			return
		default:
			strNum, err := stdout.Read(buf)
			if strNum > 0 {
				outputByte := buf[:strNum]
				_, err = dstlog.Write(outputByte)
				if err != nil {
					log.Debug().Msgf("%s 日志文件写入失败 %s", dstlog.Name(), err)
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
