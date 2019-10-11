package main

import (
	"bytes"
	"fmt"
	"github.com/andlabs/ui"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
)

type taskId int

// 事件列表
const (
	Pause = iota
	Resume
	Cancel
	Remove
	Open
	Success
	Schedule
)

var Downloader downloader

var commands = map[string]string{
	"windows": "start",
	"darwin":  "Open",
	"linux":   "xdg-Open",
}

// 下载器
type downloader struct {
	MaxRoutineNum       int
	SegSize             int64
	BufferSize          int64
	Event               chan DownloadEvent
	SavePath            string
	activeTaskNum       int32
	maxActiveTaskNum    int32
	ActiveTaskMap       map[taskId]*Task //未完成的任务
	ActiveRowToTaskId   []taskId
	CompleteTaskMap     map[taskId]*Task //已完成的任务
	CompleteRowToTaskId []taskId
	mapLock             sync.Mutex
	TaskQueue           *ItemQueue
}

// 下载器事件
type DownloadEvent struct {
	TaskId taskId
	Enum   int
}

// 初始化
func (d *downloader) Init() {
	d.Event = make(chan DownloadEvent, 1)
	d.ActiveTaskMap, d.CompleteTaskMap = make(map[taskId]*Task), make(map[taskId]*Task)
	d.ActiveRowToTaskId, d.CompleteRowToTaskId = make([]taskId, 0), make([]taskId, 0)
	d.mapLock = sync.Mutex{}
	taskQueue := &ItemQueue{
		lock: sync.RWMutex{},
	}
	d.TaskQueue = taskQueue.New()
}

// 添加任务，并为任务添加控制器
func (d *downloader) AddTask(task *Task) error {
	Downloader.mapLock.Lock()
	defer Downloader.mapLock.Unlock()
	// 创建文件
	file, err := os.OpenFile(task.SavePath, os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	_ = file.Close()
	// 创建控制器
	task.Event = &TaskEvent{
		Resume: make(chan struct{}),
		Pause:  make(chan struct{}),
		Cancel: make(chan struct{}),
	}
	task.BufferPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, d.BufferSize))
		},
	}
	// 添加任务
	Downloader.ActiveTaskMap[task.Id()] = task
	// 将任务放入等待队列
	Downloader.TaskQueue.Enqueue(task.Id())
	Downloader.Event <- DownloadEvent{
		TaskId: task.Id(),
		Enum:   Schedule,
	}
	// 维护任务列表
	Downloader.ActiveRowToTaskId = append(Downloader.ActiveRowToTaskId, task.Id())
	// 将任务添加到界面
	DpModel.RowInserted(len(Downloader.ActiveRowToTaskId) - 1)
	atomic.AddInt32(&Downloader.activeTaskNum, 1)
	return nil
}

// 获取指定任务所在行
func (d *downloader) getRow(id taskId, isActive bool) int {
	if isActive {
		for i := 0; i < len(Downloader.ActiveRowToTaskId); i++ {
			if Downloader.ActiveRowToTaskId[i] == id {
				return i
			}
		}
	} else {
		for i := 0; i < len(Downloader.CompleteRowToTaskId); i++ {
			if Downloader.CompleteRowToTaskId[i] == id {
				return i
			}
		}
	}
	return 0
}

// 暂停任务
func (d *downloader) pauseTask(id taskId) {
	task, ok := Downloader.ActiveTaskMap[id]
	if !ok {
		ui.MsgBoxError(MainWin, "错误", "任务不存在")
	}
	go task.Exit()
	task.Status = Paused
}

// 继续任务
func (d *downloader) resumeTask(id taskId) {
	task, ok := Downloader.ActiveTaskMap[id]
	if !ok {
		ui.MsgBoxError(MainWin, "错误", "任务不存在")
	}
	Downloader.TaskQueue.Enqueue(task.Id())
	Downloader.Event <- DownloadEvent{
		TaskId: task.Id(),
		Enum:   Schedule,
	}
}

// 取消任务
func (d *downloader) cancelTask(id taskId) {
	Downloader.mapLock.Lock()
	defer Downloader.mapLock.Unlock()
	task, ok := Downloader.ActiveTaskMap[id]
	if !ok {
		ui.MsgBoxError(MainWin, "错误", "任务不存在")
	}
	go task.Exit()
	_ = os.Remove(task.SavePath)
	row := Downloader.getRow(id, true)
	Downloader.ActiveRowToTaskId = append(append([]taskId{}, Downloader.ActiveRowToTaskId[:row]...),
		Downloader.ActiveRowToTaskId[row+1:]...)
	delete(Downloader.ActiveTaskMap, id)
	DpModel.RowDeleted(row)
}

// 任务完成
func (d *downloader) successTask(id taskId) {
	Downloader.mapLock.Lock()
	defer Downloader.mapLock.Unlock()
	task, _ := Downloader.ActiveTaskMap[id]
	// 将任务转移到完成界面
	row := Downloader.getRow(id, true)
	Downloader.ActiveRowToTaskId = append(append([]taskId{}, Downloader.ActiveRowToTaskId[:row]...),
		Downloader.ActiveRowToTaskId[row+1:]...)
	delete(Downloader.ActiveTaskMap, id)
	_ = task.file.Close()
	task.file = nil
	fmt.Printf("退出任务:%d", id)
	Downloader.CompleteTaskMap[id] = task
	Downloader.CompleteRowToTaskId = append(Downloader.CompleteRowToTaskId, id)
	go DpModel.RowDeleted(row)
	go CpModel.RowInserted(len(Downloader.CompleteRowToTaskId) - 1)
}

// 打开文件
func (d *downloader) openTask(id taskId) {
	task, ok := Downloader.CompleteTaskMap[id]
	if !ok {
		ui.MsgBoxError(MainWin, "错误", "任务不存在")
		return
	}
	run, ok := commands[runtime.GOOS]
	if !ok {
		ui.MsgBoxError(MainWin, "错误", fmt.Sprintf("don't know how to Open things on %s platform", runtime.GOOS))
		return
	}
	cmd := exec.Command("cmd", "/C", run, task.SavePath)
	err := cmd.Start()
	if err != nil {
		ui.MsgBoxError(MainWin, "错误", fmt.Sprintf("文件打开失败，错误原因: %v", err))
		return
	}
}

// 删除已完成任务
func (d *downloader) removeTask(id taskId) {
	task, ok := Downloader.CompleteTaskMap[id]
	if !ok {
		ui.MsgBoxError(MainWin, "错误", "任务不存在")
	}
	row := Downloader.getRow(id, false)
	Downloader.CompleteRowToTaskId = append(append([]taskId{}, Downloader.CompleteRowToTaskId[:row]...),
		Downloader.CompleteRowToTaskId[row+1:]...)
	delete(Downloader.CompleteTaskMap, task.Id())
	CpModel.RowDeleted(row)
	window := ui.NewWindow("提示", 200, 200, false)
	window.SetMargined(true)
	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	window.SetChild(vbox)
	vbox.Append(ui.NewLabel("是否删除文件?"), false)
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	vbox.Append(hbox, false)
	sure := ui.NewButton("是")
	not := ui.NewButton("否")
	hbox.Append(sure, false)
	hbox.Append(not, false)
	sure.OnClicked(func(button *ui.Button) {
		_ = os.Remove(task.SavePath)
		window.Destroy()
	})
	not.OnClicked(func(button *ui.Button) {
		window.Destroy()
	})
	window.Show()
}

// if not exist return false else return true
func (d *downloader) FileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// 任务调度
func (d *downloader) Schedule() {
	if d.activeTaskNum <= d.maxActiveTaskNum && !d.TaskQueue.IsEmpty() {
		id := (*d.TaskQueue.Dequeue()).(taskId)
		task, ok := d.ActiveTaskMap[id]
		if !ok {
			return
		}
		err := task.Start()
		if err != nil {
			return
		}
	}
}

// 监控事件
func (d *downloader) ListenEvent() {
	for {
		select {
		case event := <-d.Event:
			switch event.Enum {
			case Pause:
				MainWin.Disable()
				d.pauseTask(event.TaskId)
				break
			case Resume:
				d.resumeTask(event.TaskId)
				break
			case Cancel:
				d.cancelTask(event.TaskId)
				break
			case Remove:
				d.removeTask(event.TaskId)
				break
			case Open:
				d.openTask(event.TaskId)
				break
			case Success:
				d.successTask(event.TaskId)
				break
			case Schedule:
				d.Schedule()
				break
			}
		}
	}
}
