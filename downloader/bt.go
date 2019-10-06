package downloader

import (
	"fmt"
)

var (
	taskList       = make(map[int]chan SegMent)
	activeTaskList = make(map[int]chan struct{})
)

type bt struct {
	id   int
	task *Task
}

func startBT(fileId int) {
	seglen := len(main.seg[fileId])
	for i := 0; i < seglen; i++ {
		fileInfo := main.readFileMap(fileId)
		select {
		case <-fileInfo.Exit:
			fmt.Println("退出BT")
			return
		case <-fileInfo.Pause:
			fmt.Println("暂停BT")
			select {
			case <-fileInfo.Continue:
				fmt.Println("继续BT")
			}
		default:
			segment := main.readSegIndex(fileId, i)
			if segment.Complete {
				continue
			}
			if segment.Cache != nil {
				fileInfo.FileChan <- i
				continue
			}
		LOOP:
			if len(activeTaskList[fileId]) >= maxTaskSize {
				goto LOOP
			}
			taskList[fileId] <- segment
		}
	}
}

func startTask(fileId int) {
	for {
		fileInfo := main.readFileMap(fileId)
		select {
		case <-fileInfo.Exit:
			fmt.Println("退出Task")
			return
		case <-fileInfo.Pause:
			fmt.Println("暂停Task")
			select {
			case <-fileInfo.Continue:
				fmt.Println("继续Task")
			}
		default:
			segment := <-taskList[fileId]
			fmt.Println(segment.Index)
			// 从无缓冲队列接受segment，并发送到活动队列
			activeTaskList[fileId] <- struct{}{}
			go func(main.SegMent) {
				request := main.getRequest(segment.Url, segment.Start, segment.End)
				main.send(request, segment, fileId)
				<-activeTaskList[fileId]
				return
			}(segment)
		}
	}
}
