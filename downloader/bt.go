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
