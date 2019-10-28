package download

import (
	"downloader/common"
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

var (
	Download Downloader
)

var commands = map[string]string{
	"windows": "start",
	"darwin":  "Open",
	"linux":   "xdg-Open",
}

// 下载器
type Downloader struct {
	MaxRoutineNum    int                `json:"maxRoutineNum"`
	SegSize          int64              `json:"-"`
	BufferSize       int64              `json:"-"`
	Event            chan DownloadEvent `json:"-"`
	SavePath         string             `json:"savePath"`
	activeTaskNum    int32
	MaxActiveTaskNum int32            `json:"-"`
	ActiveTaskMap    map[string]*Task `json:"-"` //未完成的任务
	CompleteTaskMap  map[string]*Task `json:"-"` //已完成的任务
	mapLock          sync.Mutex
	TaskQueue        *common.ItemQueue `json:"-"`
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
	taskQueue := &common.ItemQueue{}
	d.TaskQueue = taskQueue.New()
	// 查询任务
	taskrow, err := common.DB.Query("select * from task")
	if err == nil {
		for taskrow.Next() {
			task := &Task{client: common.NewClient()}
			_ = taskrow.Scan(&task.Id, &task.renewal, &task.Status, &task.FileLength, &task.finalLink, &task.FileName, &task.SavePath)
			quit := make(chan struct{})
			go func(id string) {
				segrow, err := common.DB.Query("select * from segment where task_id =?", id)
				if err == nil {
					for segrow.Next() {
						segment := &SegMent{}
						_ = segrow.Scan(id, &segment.start, &segment.end, &segment.finish)
						task.completedLock.Lock()
						task.completed = append(task.completed, segment)
						task.completedLock.Unlock()
					}
				}
				quit <- struct{}{}
			}(task.Id)
			if task.Status == common.INCOMPLETE {
				task.Status = Paused
				d.ActiveTaskMap[task.Id] = task
			}
			if task.Status == common.SUCCESS {
				task.client = nil
				task.Status = Over
				d.CompleteTaskMap[task.Id] = task
			}
			if task.Status == common.ERRORED {
				task.Status = Errored
				d.ActiveTaskMap[task.Id] = task
			}
			<-quit
		}
	}
}

// 添加任务
func (d *Downloader) AddTask(fileInfo common.FileInfo, client *http.Client) (string, error) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	//if !common.FileExist(fileInfo.SavePath + fileInfo.FileName) {
	//	// 创建文件
	//	file, err := os.OpenFile(fileInfo.SavePath+fileInfo.FileName, os.O_CREATE, 0644)
	//	if err != nil {
	//		return "", err
	//	}
	//	_ = file.Close()
	//}
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
	common.DBLock.Lock()
	_, _ = common.TaskInsert.Exec(task.Id, task.renewal, common.INCOMPLETE, task.FileLength, task.finalLink, task.FileName, task.SavePath)
	common.DBLock.Unlock()
	png, _ := os.OpenFile("./resources/app/tmp/"+task.Id+".png", os.O_CREATE, 0644)
	_ = png.Close()
	// 添加任务
	Download.ActiveTaskMap[task.Id] = task
	// 将任务放入等待队列
	Download.TaskQueue.Enqueue(task.Id)
	Download.Event <- DownloadEvent{
		TaskId: task.Id,
		Enum:   Schedule,
	}
	return task.Id, nil
}

// 暂停任务
func (d *Downloader) PauseTask(id string) {
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
	}
	if i, ok := d.TaskQueue.Contains(id); ok {
		d.TaskQueue.RemoveItem(i)
	}
	if len(task.bts) != 0 {
		close(task.btCancel)
	}
	task.Status = Paused
	Download.Event <- DownloadEvent{
		TaskId: "",
		Enum:   Schedule,
	}
}

// 继续任务
func (d *Downloader) ResumeTask(id string) {
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
		return
	}
	Download.TaskQueue.Enqueue(task.Id)
	Download.Event <- DownloadEvent{
		TaskId: "",
		Enum:   Schedule,
	}
	task.Status = Waiting
	common.DBLock.Lock()
	_, _ = common.TaskUpdate.Exec(common.INCOMPLETE, task.Id)
	common.DBLock.Unlock()
}

// 取消任务
func (d *Downloader) CancelTask(id string) {
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	_ = os.Remove("./resources/app/tmp/" + id + ".png")
	task, ok := Download.ActiveTaskMap[id]
	if !ok {
		return
	}
	common.DBLock.Lock()
	_, _ = common.TaskDelete.Exec(task.Id)
	common.DBLock.Unlock()
	common.DBLock.Lock()
	_, _ = common.SegDelete.Exec(task.Id)
	common.DBLock.Unlock()
	_ = os.Remove(task.SavePath)
	if len(task.bts) != 0 {
		close(task.btCancel)
	}
	task.file = nil
	_ = os.Remove(task.SavePath)
	delete(Download.ActiveTaskMap, id)
}

// 任务完成
func (d *Downloader) successTask(id string) {
	// 维护任务列表
	atomic.AddInt32(&Download.activeTaskNum, -1)
	Download.mapLock.Lock()
	defer Download.mapLock.Unlock()
	task, _ := Download.ActiveTaskMap[id]
	// 将任务转移到完成界面
	delete(Download.ActiveTaskMap, id)
	_ = task.file.Close()
	task.file = nil
	fmt.Printf("退出任务:%s", id)
	Download.CompleteTaskMap[id] = task
	common.DBLock.Lock()
	_, _ = common.TaskUpdate.Exec(common.SUCCESS, task.Id)
	common.DBLock.Unlock()
	common.DBLock.Lock()
	_, _ = common.SegDelete.Exec(task.Id)
	common.DBLock.Unlock()
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
func (d *Downloader) RemoveTask(id string) {
	_ = os.Remove("./resources/app/tmp/" + id + ".png")
	task, ok := Download.CompleteTaskMap[id]
	if !ok {
		return
	}
	common.DBLock.Lock()
	_, _ = common.TaskDelete.Exec(task.Id)
	common.DBLock.Unlock()
	common.DBLock.Lock()
	_, _ = common.SegDelete.Exec(task.Id)
	common.DBLock.Unlock()
	delete(Download.CompleteTaskMap, task.Id)
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
			d.TaskQueue.Enqueue(id)
			return
		}
		// 维护任务列表
		atomic.AddInt32(&Download.activeTaskNum, 1)
	}
}

// 退出前的处理
func (d *Downloader) BeforeExit() {
	group := sync.WaitGroup{}
	for _, task := range d.ActiveTaskMap {
		if task.Status == Downloading {
			group.Add(1)
			go func() {
				task.Exit()
				group.Done()
			}()
		}
	}
	group.Wait()
}

// 监控事件
func (d *Downloader) ListenEvent() {
	for {
		select {
		case event := <-d.Event:
			go func() {
				switch event.Enum {
				case Pause:
					go d.PauseTask(event.TaskId)
					break
				case Resume:
					go d.ResumeTask(event.TaskId)
					break
				case Cancel:
					go d.CancelTask(event.TaskId)
					break
				case Remove:
					go d.RemoveTask(event.TaskId)
					break
				case Open:
					go d.openTask(event.TaskId)
					break
				case Success:
					go d.successTask(event.TaskId)
					break
				case Schedule:
					go d.Schedule()
					break
				}
			}()
		}
	}
}
