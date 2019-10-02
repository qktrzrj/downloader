package main

import (
	"yan.com/downloader/models"
)

const (
	maxTaskSize int = 65
)

var (
	taskList       = make(map[int]chan models.SegMent)
	activeTaskList = make(map[int]chan struct{})
)

type task struct {
	segIndex int
	active   bool
	cache    []byte
}

func startBT(fileId int) {
	for i := 0; i < len(seg[fileId]); i++ {
		select {
		case <-fileMap[fileId].Exit:
			return
		default:
			taskList[fileId] <- seg[fileId][i]
		}
	}
}

func startTask(fileId int) {
	for {
		select {
		case <-fileMap[fileId].Exit:
			return
		default :
			segment := <-taskList[fileId]
			// 从无缓冲队列接受segment，并发送到活动队列
			activeTaskList[fileId] <- struct{}{}
			go func(models.SegMent) {
				request := getRequest(segment.Url, segment.Start, segment.End)
				send(request, segment, fileId)
				<-activeTaskList[fileId]
				return
			}(segment)
		}
	}
}
