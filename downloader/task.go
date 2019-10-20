package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// 任务状态
const (
	Waiting = iota + 1
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
	Status            int  `json:"status"` //下载状态
	fileLength        int64
	DownloadCount     int64   `json:"downloadCount"` // 已下载片段数
	RemainingTime     float64 `json:"remainingTime"` //剩余时间
	Url               string
	finalLink         string
	file              *os.File
	FileName          string
	SavePath          string
	undistributed     []*SegMent // 尚未分配的片段
	undistributedLock sync.Mutex
	completed         []*SegMent
	completedLock     sync.Mutex
	btNum             int32
	bts               map[int]*bt
	btCancel          chan struct{}
	btLock            sync.Mutex
	speedCountChan    chan struct{}
	SpeedCount        float64 `json:"speedCount"`
	Event             *TaskEvent
	BufferPool        *sync.Pool
	client            *http.Client
}

func (task *Task) Id() TaskId {
	id := TaskId(task.id)
	return id
}

// 任务初始化
func (task *Task) init() (err error) {
	task.file, err = os.OpenFile(task.SavePath+task.FileName, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("打开本地文件错误:%w", err)
	}
	task.btCancel, task.speedCountChan = make(chan struct{}), make(chan struct{})
	// 创建控制器
	task.Event = &TaskEvent{
		Resume: make(chan struct{}),
		Pause:  make(chan struct{}),
		Cancel: make(chan struct{}),
	}
	// 建立缓存
	task.BufferPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, Download.SegSize))
		},
	}
	// 初始化bt
	task.bts = make(map[int]*bt)
	task.undistributed = append([]*SegMent{}, &SegMent{
		start: 0,
		end:   task.fileLength,
	})
	task.DownloadCount = 0
	if len(task.completed) > 0 {
		for _, segment := range task.completed {
			task.undistributed = task.removeSeg(task.undistributed, segment)
			task.DownloadCount += Download.SegSize
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
		for i := 0; i < Download.MaxRoutineNum; i++ {
			task.bts[i] = &bt{
				id:   i,
				task: task,
			}
			go func(bt *bt) {
				bt.start()
				task.btLock.Lock()
				delete(task.bts, bt.id)
				task.btLock.Unlock()
				log.Println(fmt.Sprintf("task %d, worker %d exit", task.id, bt.id))
				if len(task.bts) == 0 {
					if task.Status == Paused {
						log.Println(fmt.Sprintf("任务:%d 暂停成功！", task.Id()))
					}
					if task.Status == Downloading {
						task.Status = Over
						Download.Event <- DownloadEvent{
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
		preDownCount := task.DownloadCount
		select {
		case <-task.speedCountChan:
			task.SpeedCount = 0
			return
		case <-t:
			task.SpeedCount = (float64(task.DownloadCount) - float64(preDownCount)) / 1024
			task.RemainingTime = (float64(task.fileLength-task.DownloadCount) / 1024) / task.SpeedCount
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
	if segment.end-segment.start+1 > Download.SegSize {
		seg1 := &SegMent{
			start:  segment.start,
			end:    segment.start + Download.SegSize,
			finish: segment.start,
		}
		seg2 := &SegMent{
			start:  seg1.end,
			end:    segment.end,
			finish: seg1.end,
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

// 写入文件
func (task *Task) writeToDisk(segment *SegMent, buffer *bytes.Buffer) error {
	seek, err := task.file.Seek(segment.start, io.SeekStart)
	if err != nil {
		return err
	}
	if seek != segment.start {
		return errors.New("文件操作失败")
	}
	write, err := buffer.WriteTo(task.file)
	if err != nil {
		return err
	}
	if write != segment.finish-segment.start {
		return errors.New("文件写入失败")
	}
	task.file.Sync()
	task.completedLock.Lock()
	task.completed = append(task.completed, segment)
	task.completedLock.Unlock()
	return nil
}

// 除去指定长度段的segment
func (task *Task) removeSeg(seg []*SegMent, segment *SegMent) []*SegMent {
	for index := 0; index < len(seg); index++ {
		segIn := seg[index]
		if segIn.start <= segment.start && segIn.end >= segment.end {
			if segIn.start == segment.start && segIn.end == segment.end {
				return append(seg[:index], seg[index+1:]...)
			}
			if segIn.start == segment.start {
				segIn.start = segment.end + 1
				return seg
			}
			if segIn.end == segment.end {
				segIn.end = segment.start - 1
				return seg
			}
			segInsert := &SegMent{
				start: segment.end + 1,
				end:   segIn.end,
			}
			segIn.end = segment.start - 1
			real := append([]*SegMent{}, seg[index+1:]...)
			return append(append(seg[:index+1], segInsert), real...)
		}
	}
	return seg
}
