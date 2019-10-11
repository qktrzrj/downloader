package main

import (
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"sync/atomic"
)

var (
	taskList       = make(map[int]chan SegMent)
	activeTaskList = make(map[int]chan struct{})
)

type bt struct {
	id   int
	task *Task
}

func (bt *bt) start() {
	errNum := 0
	for {
		if errNum >= 3 {
			log.Println(fmt.Sprintf("task %d, worker %d error", bt.task.id, bt.id))
			bt.task.Status = Errored
			go bt.task.Exit()
			DpModel.RowChanged(Downloader.getRow(bt.task.Id(), true))
			break
		}
		select {
		case <-bt.task.btCancel:
			return
		default:
			segment := bt.task.getSeg()
			if segment == nil {
				return
			}
			err := bt.downSeg(segment)
			if err != nil {
				log.Println(fmt.Sprintf("task %d, worker %d segment start %d end %d error",
					bt.task.id, bt.id, segment.Start, segment.End))
				errNum++
				bt.task.segErr(segment)
				continue
			}
			atomic.AddInt64(&bt.task.downloadCount, Downloader.SegSize)
		}
	}
}

func (bt *bt) downSeg(segment *SegMent) error {
	req := getRequest(bt.task.finalLink, int(segment.Start), int(segment.End))
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	//defer
	err := bt.task.client.Do(req, resp)
	if err != nil {
		return err
	}
	if resp.StatusCode() != fasthttp.StatusOK && resp.StatusCode() != fasthttp.StatusPartialContent {
		return errors.New("下载错误")
	}
	_, _ = bt.task.file.Seek(segment.Start, io.SeekStart)
	err = resp.BodyWriteTo(bt.task.file)
	if err != nil {
		return err
	}
	bt.task.file.Sync()
	bt.task.writeToDiskLock.Lock()
	bt.task.writeToDisk = append(bt.task.writeToDisk, segment)
	bt.task.writeToDiskLock.Unlock()

	return nil
}
