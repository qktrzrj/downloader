package downloader

import (
	"bytes"
	ui2 "downloader/ui"
	"fmt"
	"github.com/andlabs/ui"
	"os"
	"os/exec"
	"runtime"
	"sync"
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
	BufferPool          sync.Pool
	activeTaskNum       int
	maxActiveTaskNum    int
	ActiveTaskMap       map[taskId]*Task //未完成的任务
	ActiveRowToTaskId   [][1]int
	CompleteTaskMap     map[taskId]*Task //已完成的任务
	CompleteRowToTaskId [][1]int
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
	d.ActiveRowToTaskId, d.CompleteRowToTaskId = make([][1]int, 0), make([][1]int, 0)
	taskQueue := &ItemQueue{
		lock: sync.RWMutex{},
	}
	d.TaskQueue = taskQueue.New()
	d.BufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, d.BufferSize))
		},
	}
}

// 添加任务，并为任务添加控制器
func (d *downloader) AddTask(task *Task) error {
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
	// 添加任务
	Downloader.ActiveTaskMap[task.Id()] = task
	// 将任务放入等待队列
	Downloader.TaskQueue.Enqueue(task.Id())
	// 维护任务列表
	rowToId := [1]int{task.id}
	Downloader.ActiveRowToTaskId = append(Downloader.ActiveRowToTaskId, rowToId)
	// 将任务添加到界面
	ui2.DpModel.RowInserted(len(Downloader.ActiveRowToTaskId) - 1)
	return nil
}

// 暂停任务
func (d *downloader) pauseTask(id taskId) {
	task, ok := Downloader.ActiveTaskMap[id]
	if !ok {
		ui.MsgBoxError(ui2.MainWin, "错误", "任务不存在")
	}
	task.Event.Pause <- struct{}{}
}

// 继续任务
func (d *downloader) resumeTask(id taskId) {
	task, ok := Downloader.ActiveTaskMap[id]
	if !ok {
		ui.MsgBoxError(ui2.MainWin, "错误", "任务不存在")
	}
	task.Event.Resume <- struct{}{}
}

// 取消任务
func (d *downloader) cancelTask(id taskId) {
	task, ok := Downloader.ActiveTaskMap[id]
	if !ok {
		ui.MsgBoxError(ui2.MainWin, "错误", "任务不存在")
	}
	task.Event.Cancel <- struct{}{}
	delete(Downloader.ActiveTaskMap, id)
}

// 任务完成
//func (d *downloader) successTask(id taskId) {
//	task, _ := Downloader.ActiveTaskMap[id]
//	// 将任务转移到完成界面
//
//}

// 打开文件
func (d *downloader) openTask(id taskId) {
	task, ok := Downloader.CompleteTaskMap[id]
	if !ok {
		ui.MsgBoxError(ui2.MainWin, "错误", "任务不存在")
	}
	run, ok := commands[runtime.GOOS]
	if !ok {
		ui.MsgBoxError(ui2.MainWin, "错误", fmt.Sprintf("don't know how to Open things on %s platform", runtime.GOOS))
	}
	cmd := exec.Command(run, task.SavePath+"/"+task.FileName)
	err := cmd.Start()
	if err != nil {
		ui.MsgBoxError(ui2.MainWin, "错误", fmt.Sprintf("文件打开失败，错误原因: %v", err))
	}
}

// 删除已完成任务
func (d *downloader) removeTask(id taskId) {
	task, ok := Downloader.CompleteTaskMap[id]
	if !ok {
		ui.MsgBoxError(ui2.MainWin, "错误", "任务不存在")
	}
	delete(Downloader.CompleteTaskMap, task.Id())
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

// 任务调度
func (d *downloader) Schedule() {
	for {
		if d.activeTaskNum <= d.maxActiveTaskNum && !d.TaskQueue.IsEmpty() {
			id := (*d.TaskQueue.Dequeue()).(taskId)
			task, ok := d.ActiveTaskMap[id]
			if !ok {
				continue
			}
			err := task.Start()
			go task.listen()
			if err != nil {
				continue
			}
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
				break
			}
		}
	}
}
