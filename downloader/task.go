package downloader

import (
	"downloader/ui"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"log"
	"os"
	"sync"
	"time"
)

// 任务状态
const (
	Waiting = iota
	Downloading
	Over
	Paused
	Errored
)

// 任务事件
type TaskEvent struct {
	Resume chan struct{}
	Pause  chan struct{}
	Cancel chan struct{}
	Error  chan struct{}
}

type Task struct {
	id                int
	renewal           bool // 是否支持断点续传
	Status            int  //下载状态
	fileLength        int64
	downloadCount     int64 // 已下载片段数
	Url               string
	finalLink         string
	file              *os.File
	FileName          string
	SavePath          string
	undistributed     []*SegMent // 尚未分配的片段
	writeToDisk       []*SegMent // 文件内容写入磁盘的情况
	undistributedLock sync.Mutex
	writeToDiskLock   sync.Mutex
	bts               map[int]*bt
	btCancel          chan struct{}
	btLock            sync.Mutex
	speedCountChan    chan struct{}
	speedCount        float64
	Event             *TaskEvent
	client            *fasthttp.HostClient
}

func (task *Task) Id() taskId {
	id := taskId(task.id)
	return id
}

// 任务初始化
func (task *Task) init() (err error) {
	task.file, err = os.OpenFile(task.SavePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("打开本地文件错误:%w", err)
	}
	task.undistributedLock, task.btLock, task.writeToDiskLock = sync.Mutex{}, sync.Mutex{}, sync.Mutex{}
	task.btCancel, task.speedCountChan = make(chan struct{}), make(chan struct{})
	// 初始化bt
	task.bts = make(map[int]*bt)
	task.undistributed = append(task.undistributed, &SegMent{
		Start: 0,
		End:   task.fileLength,
	})
	if len(task.writeToDisk) != 0 {
		for _, segment := range task.writeToDisk {
			task.removeSeg(task.undistributed, segment)
			task.downloadCount = task.downloadCount + segment.End - segment.Start + 1
		}
	}
	return
}

func (task *Task) Start() error {
	if task.Status != Waiting && task.Status != Paused {
		return errors.New("当前状态无法下载！")
	}
	task.Status = Downloading
	go func() {
		err := task.init()
		if err != nil {
			task.Status = Errored
			return
		}
		go task.speedCalculate()
		for i := 0; i < Downloader.MaxRoutineNum; i++ {
			task.bts[i] = &bt{
				id:   i,
				task: task,
			}
			go func(bt *bt) {
				bt.start()
				task.btLock.Lock()
				defer task.btLock.Unlock()
				delete(task.bts, bt.id)
				log.Println(fmt.Sprintf("task %s, worker %d exit", task.id, bt.id))
				if len(task.bts) == 0 {
					go task.Exit()
					task.Status = Over
					Downloader.Event <- DownloadEvent{
						TaskId: task.Id(),
						Enum:   Success,
					}
				}
			}(task.bts[i])
			go task.listen()
		}
	}()
	return nil
}

// 退出任务
func (task *Task) Exit() {
	close(task.speedCountChan)
	_ = task.file.Close()
}

// 事件监听
func (task *Task) listen() {
	for {
		select {
		case <-task.Event.Pause:
			go task.Exit()
			close(task.btCancel)
			task.Status = Paused
		case <-task.Event.Resume:
			_ = task.Start()
		case <-task.Event.Cancel:
			go task.Exit()
			if len(task.bts) != 0 {
				close(task.btCancel)
			}
		case <-task.Event.Error:
			task.Status = Errored
			go task.Exit()
		}
	}
}

// 计算下载速度
func (task *Task) speedCalculate() {
	t := time.Tick(time.Second)
	for {
		preDownCount := task.downloadCount
		select {
		case <-task.speedCountChan:
			return
		case <-t:
			task.speedCount = (float64(task.downloadCount) - float64(preDownCount)) / 1024
			ui.DpModel.RowChanged(0)
		}
	}
}

// 获取片段
func (task *Task) getSeg() *SegMent {
	task.undistributedLock.Lock()
	defer task.undistributedLock.Unlock()
	length := len(task.undistributed)
	if length == 0 {
		return nil
	}
	segment := task.undistributed[0]
	if segment.End-segment.Start+1 > Downloader.SegSize {
		seg1 := &SegMent{
			Start: segment.Start,
			End:   segment.Start + Downloader.SegSize - 1,
		}
		seg2 := &SegMent{
			Start: seg1.End + 1,
			End:   segment.End,
		}
		task.undistributed[0] = seg2
		segment = seg1
	} else {
		task.undistributed = task.undistributed[1:]
	}
	return segment
}

// 下载出错，将片段放回
func (task *Task) segErr(segment *SegMent) {
	task.undistributedLock.Lock()
	defer task.undistributedLock.Unlock()
	task.undistributed = append(task.undistributed[1:], task.undistributed...)
	task.undistributed[0] = segment
}

// 下载成功，将片段放回
func (task *Task) segSuccess(segment *SegMent) {
	task.undistributedLock.Lock()
	defer task.undistributedLock.Unlock()
	task. = append(task.undistributed[1:], task.undistributed...)
	task.undistributed[0] = segment
}

// 除去指定长度段的segment
func (task *Task) removeSeg(seg []*SegMent, segment *SegMent) []*SegMent {
	for index := 0; index < len(seg); index++ {
		segIn := seg[index]
		if segIn.Start <= segment.Start && segIn.End >= segment.End {
			if segIn.Start == segment.Start && segIn.End == segment.End {
				return append(seg[:index], seg[index+1:]...)
			}
			if segIn.Start == segment.Start {
				segIn.Start = segment.End + 1
				return seg
			}
			if segIn.End == segment.End {
				segIn.End = segment.Start - 1
				return seg
			}
			segInsert := &SegMent{
				Start: segment.End + 1,
				End:   segIn.End,
			}
			segIn.End = segment.Start - 1
			real := append([]*SegMent{}, seg[index+1:]...)
			return append(append(seg[:index+1], segInsert), real...)
		}
	}
	return seg
}
