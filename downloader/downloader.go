package downloader

import (
	"fmt"
	"github.com/guonaihong/gout"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
)

type TaskId int

// 事件列表
const (
	Pause = iota + 1
	Resume
	Cancel
	Remove
	Open
	Success
	Schedule
)

var Download Downloader

var commands = map[string]string{
	"windows": "start",
	"darwin":  "Open",
	"linux":   "xdg-Open",
}

// 下载器
type Downloader struct {
	MaxRoutineNum       int
	SegSize             int64
	BufferSize          int64
	Event               chan DownloadEvent
	SavePath            string
	activeTaskNum       int32
	MaxActiveTaskNum    int32
	ActiveTaskMap       map[TaskId]*Task //未完成的任务
	ActiveRowToTaskId   []TaskId
	CompleteTaskMap     map[TaskId]*Task //已完成的任务
	CompleteRowToTaskId []TaskId
	mapLock             sync.Mutex
	TaskQueue           *ItemQueue
}

// 下载器事件
type DownloadEvent struct {
	TaskId TaskId
	Enum   int
}

// 初始化
func (d *Downloader) Init() {
	d.Event = make(chan DownloadEvent, 1)
	d.ActiveTaskMap, d.CompleteTaskMap = make(map[TaskId]*Task), make(map[TaskId]*Task)
	d.ActiveRowToTaskId, d.CompleteRowToTaskId = make([]TaskId, 0), make([]TaskId, 0)
	d.mapLock = sync.Mutex{}
	taskQueue := &ItemQueue{
		lock: sync.RWMutex{},
	}
	d.TaskQueue = taskQueue.New()
}

// 添加任务
func (d *Downloader) AddTask(fileInfo FileInfo, g *gout.Gout) (TaskId, error) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	// 创建文件
	file, err := os.OpenFile(fileInfo.SavePath, os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	_ = file.Close()
	// 创建任务
	task := &Task{
		id:         rand.Int(),
		renewal:    fileInfo.Renewal,
		Status:     Waiting,
		fileLength: fileInfo.Length,
		finalLink:  fileInfo.FinalLink,
		FileName:   fileInfo.FileName,
		SavePath:   fileInfo.SavePath,
		client:     g,
	}
	// 添加任务
	Download.ActiveTaskMap[task.Id()] = task
	// 将任务放入等待队列
	Download.TaskQueue.Enqueue(task.Id())
	Download.Event <- DownloadEvent{
		TaskId: task.Id(),
		Enum:   Schedule,
	}
	// 维护任务列表
	Download.ActiveRowToTaskId = append(Download.ActiveRowToTaskId, task.Id())
	atomic.AddInt32(&Download.activeTaskNum, 1)
	return task.Id(), nil
}

// 获取指定任务所在行
func (d *Downloader) getRow(id TaskId, isActive bool) int {
	if isActive {
		for i := 0; i < len(Download.ActiveRowToTaskId); i++ {
			if Download.ActiveRowToTaskId[i] == id {
				return i
			}
		}
	} else {
		for i := 0; i < len(Download.CompleteRowToTaskId); i++ {
			if Download.CompleteRowToTaskId[i] == id {
				return i
			}
		}
	}
	return 0
}

// 暂停任务
func (d *Downloader) pauseTask(id TaskId) {
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	go task.Exit()
	task.Status = Paused
}

// 继续任务
func (d *Downloader) resumeTask(id TaskId) {
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	Download.TaskQueue.Enqueue(task.Id())
	Download.Event <- DownloadEvent{
		TaskId: task.Id(),
		Enum:   Schedule,
	}
}

// 取消任务
func (d *Downloader) cancelTask(id TaskId) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	go task.Exit()
	task.file = nil
	_ = os.Remove(task.SavePath)
	row := Download.getRow(id, true)
	Download.ActiveRowToTaskId = append(append([]TaskId{}, Download.ActiveRowToTaskId[:row]...),
		Download.ActiveRowToTaskId[row+1:]...)
	delete(Download.ActiveTaskMap, id)
}

// 任务完成
func (d *Downloader) successTask(id TaskId) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	task, _ := Download.ActiveTaskMap[id]
	// 将任务转移到完成界面
	row := Download.getRow(id, true)
	Download.ActiveRowToTaskId = append(append([]TaskId{}, Download.ActiveRowToTaskId[:row]...),
		Download.ActiveRowToTaskId[row+1:]...)
	delete(Download.ActiveTaskMap, id)
	_ = task.file.Close()
	task.file = nil
	fmt.Printf("退出任务:%d", id)
	Download.CompleteTaskMap[id] = task
	Download.CompleteRowToTaskId = append(Download.CompleteRowToTaskId, id)
}

// 打开文件
func (d *Downloader) openTask(id TaskId) {
	task, ok := Download.CompleteTaskMap[id]
	if !ok {
		return
	}
	run, ok := commands[runtime.GOOS]
	if !ok {
		return
	}
	cmd := exec.Command("cmd", "/C", run, task.SavePath)
	err := cmd.Start()
	if err != nil {
		return
	}
}

// 删除已完成任务
func (d *Downloader) removeTask(id TaskId) {
	task, ok := Download.CompleteTaskMap[id]
	if !ok {
	}
	row := Download.getRow(id, false)
	Download.CompleteRowToTaskId = append(append([]TaskId{}, Download.CompleteRowToTaskId[:row]...),
		Download.CompleteRowToTaskId[row+1:]...)
	delete(Download.CompleteTaskMap, task.Id())
	//main.CpModel.RowDeleted(row)
	//window := resources.NewWindow("提示", 200, 200, false)
	//window.SetMargined(true)
	//vbox := resources.NewVerticalBox()
	//vbox.SetPadded(true)
	//window.SetChild(vbox)
	//vbox.Append(resources.NewLabel("是否删除文件?"), false)
	//hbox := resources.NewHorizontalBox()
	//hbox.SetPadded(true)
	//vbox.Append(hbox, false)
	//sure := resources.NewButton("是")
	//not := resources.NewButton("否")
	//hbox.Append(sure, false)
	//hbox.Append(not, false)
	//sure.OnClicked(func(button *resources.Button) {
	//	_ = os.Remove(task.SavePath)
	//	window.Destroy()
	//})
	//not.OnClicked(func(button *resources.Button) {
	//	window.Destroy()
	//})
	//window.Show()
}

// if not exist return false else return true
func (d *Downloader) FileExist(path string) bool {
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
func (d *Downloader) Schedule() {
	if d.activeTaskNum <= d.MaxActiveTaskNum && !d.TaskQueue.IsEmpty() {
		id := (*d.TaskQueue.Dequeue()).(TaskId)
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
func (d *Downloader) ListenEvent() {
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
				d.successTask(event.TaskId)
				break
			case Schedule:
				d.Schedule()
				break
			}
		}
	}
}
