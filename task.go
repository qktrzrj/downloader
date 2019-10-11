package main

import (
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

var TaskStatusMap map[int]string = map[int]string{
	Waiting:     "等待",
	Downloading: "下载中",
	Over:        "下载完成",
	Paused:      "暂停",
	Errored:     "下载出错",
}

// what can do when task stay slow status
var DpStatusMap map[int]string = map[int]string{
	Waiting:     "暂停",
	Downloading: "暂停",
	Paused:      "继续",
	Errored:     "重新下载",
}

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
	downloadCount     int64   // 已下载片段数
	remainingTime     float64 //剩余时间
	Url               string
	finalLink         string
	file              *os.File
	FileName          string
	SavePath          string
	undistributed     []*SegMent // 尚未分配的片段
	undistributedLock sync.Mutex
	writeToDisk       []*SegMent
	writeToDiskLock   sync.Mutex
	btNum             int32
	bts               map[int]*bt
	btCancel          chan struct{}
	btLock            sync.Mutex
	speedCountChan    chan struct{}
	speedCount        float64
	Event             *TaskEvent
	BufferPool        *sync.Pool
	client            *fasthttp.HostClient
}

func (task *Task) Id() taskId {
	id := taskId(task.id)
	return id
}

// 任务初始化
func (task *Task) init() (err error) {
	task.file, err = os.OpenFile(task.SavePath, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("打开本地文件错误:%w", err)
	}
	task.undistributedLock, task.btLock, task.writeToDiskLock = sync.Mutex{}, sync.Mutex{}, sync.Mutex{}
	task.btCancel, task.speedCountChan = make(chan struct{}), make(chan struct{})
	// 初始化bt
	task.bts = make(map[int]*bt)
	task.undistributed = append([]*SegMent{}, &SegMent{
		Start: 0,
		End:   task.fileLength,
	})
	task.downloadCount = 0
	if len(task.writeToDisk) > 0 {
		for _, segment := range task.writeToDisk {
			task.undistributed = task.removeSeg(task.undistributed, segment)
			task.downloadCount += Downloader.SegSize
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
		//go task.barCalcuate()
		for i := 0; i < Downloader.MaxRoutineNum; i++ {
			task.bts[i] = &bt{
				id:   i,
				task: task,
			}
			//atomic.AddInt32(&task.btNum, 1)
			go func(bt *bt) {
				bt.start()
				task.btLock.Lock()
				delete(task.bts, bt.id)
				task.btLock.Unlock()
				//atomic.AddInt32(&task.btNum, -1)
				log.Println(fmt.Sprintf("task %d, worker %d exit", task.id, bt.id))
				if len(task.bts) == 0 {
					if task.Status == Paused {
						log.Println(fmt.Sprintf("任务:%d 暂停成功！", task.Id()))
						MainWin.Enable()
						DpModel.RowChanged(Downloader.getRow(task.Id(), true))
					}
					if task.Status == Downloading {
						task.Status = Over
						Downloader.Event <- DownloadEvent{
							TaskId: task.Id(),
							Enum:   Success,
						}
					}
				}
			}(task.bts[i])
		}
	}()
	return nil
}

// 退出任务
func (task *Task) Exit() {
	if len(task.bts) != 0 {
		close(task.btCancel)
	}
	close(task.speedCountChan)
	_ = task.file.Close()
}

// 计算下载速度
func (task *Task) speedCalculate() {
	t := time.Tick(time.Second)
	for {
		preDownCount := task.downloadCount
		select {
		case <-task.speedCountChan:
			task.speedCount = 0
			return
		case <-t:
			task.speedCount = (float64(task.downloadCount) - float64(preDownCount)) / 1024
			task.remainingTime = (float64(task.fileLength-task.downloadCount) / 1024) / task.speedCount
			DpModel.RowChanged(Downloader.getRow(task.Id(), true))
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
	segment := task.undistributed[length-1]
	if segment.End-segment.Start+1 > Downloader.SegSize {
		seg1 := &SegMent{
			Start: segment.Start,
			End:   segment.Start + Downloader.SegSize - 1,
		}
		seg2 := &SegMent{
			Start: seg1.End + 1,
			End:   segment.End,
		}
		task.undistributed = task.undistributed[:length-1]
		task.undistributed = append(task.undistributed, seg2)
		segment = seg1
	} else {
		task.undistributed = task.undistributed[:length-1]
	}
	return segment
}

// 下载出错，将片段放回
func (task *Task) segErr(segment *SegMent) {
	task.undistributedLock.Lock()
	defer task.undistributedLock.Unlock()
	task.undistributed = append(task.undistributed, segment)
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
