package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestBt__downSeg(t *testing.T) {
	url := "https://www.typora.io/windows/typora-setup-x64.exe"
	Downloader = downloader{
		MaxRoutineNum:    200,
		SegSize:          100 * 1024,
		BufferSize:       0,
		SavePath:         "F:/goProject/downloader/",
		maxActiveTaskNum: 3,
	}
	Downloader.Init()
	task, err := GetFileTask(url, NewHostClient(url))
	if err != nil {
		panic(err)
	}
	task.SavePath = Downloader.SavePath + task.FileName
	// 创建文件
	file, _ := os.OpenFile(task.SavePath, os.O_CREATE, 0644)
	_ = file.Close()
	// 创建控制器
	task.Event = &TaskEvent{
		Resume: make(chan struct{}),
		Pause:  make(chan struct{}),
		Cancel: make(chan struct{}),
	}
	// 添加任务
	Downloader.ActiveTaskMap[task.Id()] = task
	// 将任务放入等待队列
	Downloader.TaskQueue.Enqueue(task.Id())
	// 维护任务列表
	Downloader.ActiveRowToTaskId = append(Downloader.ActiveRowToTaskId, task.Id())
	_ = task.Start()
	fmt.Println(time.Now())
	select {
	case <-Downloader.Event:
		fmt.Println(time.Now())
		task.file.Close()
		return
	}
}

func Test__filePath(t *testing.T) {
	current, _ := user.Current()
	fmt.Println(current.HomeDir)
	fileInfo, _ := os.Stat(current.HomeDir)
	s, _ := filepath.Abs(filepath.Dir(fileInfo.Name()))
	fmt.Println(s)
	run, _ := commands[runtime.GOOS]
	cmd := exec.Command("cmd", "/C", run, s)
	fmt.Println(cmd.String())
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
}
