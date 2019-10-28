package download

import (
	"bufio"
	"bytes"
	"downloader/common"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
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

// 任务事件
type TaskEvent struct {
	Resume chan struct{}
	Pause  chan struct{}
	Cancel chan struct{}
	Error  chan struct{}
}

type Task struct {
	Id                string  `json:"id"`
	renewal           bool    // 是否支持断点续传
	Status            int     `json:"status"` //下载状态
	FileLength        int64   `json:"filelength"`
	DownloadCount     int64   `json:"downloadCount"` // 已下载片段数
	RemainingTime     float64 `json:"remainingTime"` //剩余时间
	Url               string  `json:"-"`
	finalLink         string
	file              *os.File
	FileName          string     `json:"-"`
	SavePath          string     `json:"-"`
	undistributed     []*SegMent // 尚未分配的片段
	undistributedLock sync.Mutex
	completed         []*SegMent
	completedLock     sync.Mutex
	btControl         int32
	bts               map[int]*bt
	btCancel          chan struct{}
	btLock            sync.Mutex
	speedCountChan    chan struct{}
	SpeedCount        float64    `json:"speedCount"`
	Event             *TaskEvent `json:"-"`
	BufferPool        *sync.Pool `json:"-"`
	client            *http.Client
	Conn              *websocket.Conn `json:"-"`
}

// 任务初始化
func (task *Task) init() (err error) {
	if !task.renewal {
		_ = os.Remove(task.SavePath + task.FileName)
	}
	task.file, err = os.OpenFile(task.SavePath+task.FileName, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
	if err != nil {
		return fmt.Errorf("打开本地文件错误:%w", err)
	}

	if task.file == nil {
		return errors.New("打开本地文件错误:文件不存在")
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
			return bytes.NewBuffer(make([]byte, 0, common.BufferSize))
		},
	}
	// 初始化bt
	task.bts = make(map[int]*bt)
	task.undistributed = append([]*SegMent{}, &SegMent{
		start: 0,
		end:   task.FileLength - 1,
	})
	task.DownloadCount = 0
	if len(task.completed) > 0 && task.renewal {
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
	//task.btControl = false
	err := task.init()
	if err != nil {
		task.Status = Errored
		return err
	}
	go task.speedCalculate()
	go func() {
		routine := Download.MaxRoutineNum
		if task.FileLength <= 0 {
			routine = 1
		} else {
			size := int(task.FileLength / Download.SegSize)
			if size < routine {
				routine = size
			}
		}
		task.btControl = int32(routine)
		for i := 0; i < routine; i++ {
			//task.btLock.Lock()
			task.bts[i] = &bt{
				id:      i,
				task:    task,
				request: common.GetRequest(task.finalLink),
			}
			go func(bt *bt) {
				bt.start()
				//task.btLock.Lock()
				log.Println(fmt.Sprintf("task %s, worker %d exit", task.Id, bt.id))
				if atomic.AddInt32(&task.btControl, -1) == 0 {
					task.bts = nil
					if task.Status == Paused {
						log.Println(fmt.Sprintf("任务:%s 暂停成功！", task.Id))
					}
					if task.Status == Errored {
						common.DBLock.Lock()
						_, _ = common.TaskUpdate.Exec(common.ERRORED, task.Id)
						common.DBLock.Unlock()
						log.Println(fmt.Sprintf("任务:%s 下载失败！", task.Id))
					}
					if task.Status == Downloading {
						task.Status = Over
						fmt.Println("下载完成")
						Download.Event <- DownloadEvent{
							TaskId: task.Id,
							Enum:   Success,
						}
					}
					time.Sleep(time.Second)
					go task.Exit()
				}
				//task.btLock.Unlock()
			}(task.bts[i])
			//task.btLock.Unlock()
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
	if task.Conn != nil {
		_ = task.Conn.Close()
	}
	_ = task.file.Close()
}

// 计算下载速度
func (task *Task) speedCalculate() {
	t := time.Tick(time.Second)
	path := task.SavePath + task.FileName
	for {
		preDownCount := task.DownloadCount
		select {
		case <-task.speedCountChan:
			task.SpeedCount = 0
			return
		case <-t:
			task.DownloadCount = common.GetFileSize(path)
			task.SpeedCount, _ = decimal.NewFromFloat(float64(task.DownloadCount) - float64(preDownCount)).Div(
				decimal.NewFromFloat(1024)).Float64()
			if decimal.NewFromFloat(task.SpeedCount).GreaterThan(decimal.NewFromFloat(0)) {
				task.RemainingTime, _ = (decimal.NewFromFloat(float64(task.FileLength - task.DownloadCount)).Div(
					decimal.NewFromFloat(1024))).Div(decimal.NewFromFloat(task.SpeedCount)).Float64()
			}
			if task.Conn != nil {
				_ = task.Conn.WriteJSON(task)
			}
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
	if segment.end-segment.start+1 > Download.SegSize && task.renewal {
		seg1 := &SegMent{
			start:  segment.start,
			end:    segment.start + Download.SegSize - 1,
			finish: segment.start,
		}
		seg2 := &SegMent{
			start:  seg1.end + 1,
			end:    segment.end,
			finish: seg1.end + 1,
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
func (task *Task) writeToDisk(segment *SegMent, buffer *bytes.Buffer) (err error) {
	_, err = task.file.Seek(segment.finish, io.SeekStart)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(task.file)
	//if seek != segment.start {
	//	return errors.New("文件操作失败")
	//}
	//l, err := writer.Write(buffer.Bytes())
	l, err := buffer.WriteTo(writer)
	if err != nil {
		return
	}
	segment.finish = segment.start + l - 1 // 片段写入磁盘偏移量
	seg := &SegMent{
		start:  segment.start,
		end:    segment.finish,
		finish: segment.finish,
	}
	if segment.finish != segment.end && task.renewal && segment.end > 0 {
		segment.start = segment.finish + 1
		task.segErr(segment)
	}
	task.file.Sync()
	if !task.renewal {
		return
	}
	task.completedLock.Lock()
	task.completed = append(task.completed, seg)
	common.DBLock.Lock()
	_, _ = common.SegInsert.Exec(task.Id, seg.start, seg.end, seg.finish)
	common.DBLock.Unlock()
	task.completedLock.Unlock()
	return
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
			reals := append([]*SegMent{}, seg[index+1:]...)
			return append(append(seg[:index+1], segInsert), reals...)
		}
	}
	return seg
}
