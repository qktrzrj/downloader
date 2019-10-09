package main

import (
	"github.com/valyala/fasthttp"
	"io"
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
			bt.task.Status = Errored
			return
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
				errNum++
				bt.task.segErr(segment)
				continue
			}
		}
	}
}

func (bt *bt) downSeg(segment *SegMent) error {
	req := getRequest(bt.task.finalLink, int(segment.Start), int(segment.End))
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	defer bt.task.file.Sync()
	err := bt.task.client.Do(req, resp)
	if err != nil {
		return err
	}
	_, _ = bt.task.file.Seek(segment.Start, io.SeekStart)
	err = resp.BodyWriteTo(bt.task.file)
	if err != nil {
		return err
	}
	return nil
}
