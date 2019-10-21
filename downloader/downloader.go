package downloader

import (
	"downloader/util"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"
)

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
	MaxRoutineNum    int
	SegSize          int64
	BufferSize       int64
	Event            chan DownloadEvent
	SavePath         string
	activeTaskNum    int32
	MaxActiveTaskNum int32
	ActiveTaskMap    map[string]*Task //未完成的任务
	CompleteTaskMap  map[string]*Task //已完成的任务
	mapLock          sync.Mutex
	TaskQueue        *ItemQueue
}

// 下载器事件
type DownloadEvent struct {
	TaskId string `json:"taskid"`
	Enum   int    `json:"enum"`
}

// 初始化
func (d *Downloader) Init() {
	d.Event = make(chan DownloadEvent, 1)
	d.ActiveTaskMap, d.CompleteTaskMap = make(map[string]*Task), make(map[string]*Task)
	d.mapLock = sync.Mutex{}
	taskQueue := &ItemQueue{
		lock: sync.RWMutex{},
	}
	d.TaskQueue = taskQueue.New()
}

// 添加任务
func (d *Downloader) AddTask(fileInfo util.FileInfo, client *http.Client) (string, error) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	if d.FileExist(fileInfo.SavePath + fileInfo.FileName) {
		_ = os.Remove(fileInfo.SavePath + fileInfo.FileName)
	}
	// 创建文件
	file, err := os.OpenFile(fileInfo.SavePath+fileInfo.FileName, os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	_ = file.Close()
	// 创建任务
	task := &Task{
		Id:         uuid.NewV4().String(),
		renewal:    fileInfo.Renewal,
		Status:     Waiting,
		FileLength: fileInfo.Length,
		finalLink:  fileInfo.FinalLink,
		FileName:   fileInfo.FileName,
		SavePath:   fileInfo.SavePath,
		client:     client,
		Conn:       nil,
	}
	// 添加任务
	Download.ActiveTaskMap[task.Id] = task
	// 将任务放入等待队列
	Download.TaskQueue.Enqueue(task.Id)
	Download.Event <- DownloadEvent{
		TaskId: task.Id,
		Enum:   Schedule,
	}
	// 维护任务列表
	atomic.AddInt32(&Download.activeTaskNum, 1)
	return task.Id, nil
}

// 暂停任务
func (d *Downloader) pauseTask(id string) {
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	go task.Exit()
	task.Status = Paused
}

// 继续任务
func (d *Downloader) resumeTask(id string) {
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	Download.TaskQueue.Enqueue(task.Id)
	Download.Event <- DownloadEvent{
		TaskId: task.Id,
		Enum:   Schedule,
	}
}

// 取消任务
func (d *Downloader) cancelTask(id string) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	go task.Exit()
	task.file = nil
	_ = os.Remove(task.SavePath)
	delete(Download.ActiveTaskMap, id)
}

// 任务完成
func (d *Downloader) successTask(id string) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	task, _ := Download.ActiveTaskMap[id]
	// 将任务转移到完成界面
	delete(Download.ActiveTaskMap, id)
	_ = task.file.Close()
	task.file = nil
	fmt.Printf("退出任务:%d", id)
	Download.CompleteTaskMap[id] = task
}

// 打开文件
func (d *Downloader) openTask(id string) {
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
func (d *Downloader) removeTask(id string) {
	task, ok := Download.CompleteTaskMap[id]
	if !ok {
	}
	delete(Download.CompleteTaskMap, task.Id)
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
		id := (*d.TaskQueue.Dequeue()).(string)
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
