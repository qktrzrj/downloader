package downloader

import (
	"downloader"
	"downloader/ui"
	"fmt"
	"time"
)

var schedule = make(chan int)
var sch = make(chan struct{})
var filenum int

func scheduler() {
	for {
	LOOP:
		select {
		case fileId := <-schedule:
			if filenum >= 3 {
				main.fileQueue.Enqueue(fileId)
				sch <- struct{}{}
				goto LOOP
			}
			filenum++
			go main.download(fileId)
		}
	}
}

func getRate(fileId int, t time.Time) {
	for {
		fileInfo := main.readFileMap(fileId)
		tablemodel := ui.model[0]
		select {
		case <-fileInfo.Exit:
			fmt.Println("退出Rate")
			main.group.Done()
			return
		case <-fileInfo.Pause:
			fmt.Println("暂停Rate")
			select {
			case <-fileInfo.Continue:
				fmt.Println("继续Rate")
			}
		default:
			// 获取已写入块总和大小
			size := main.getFileSize(fileId)
			main.barList[fileId] = getPercent(size, int64(fileInfo.Length))
			tablemodel.RowChanged(fileInfo.Row)
			//result, _ := strconv.Atoi(percent)
			//str := "working " + percent + "%" + "[" + bar(result, 100) + "] " + " " +
			//	fmt.Sprintf("%.f", getCurrentSize(t)) + "s"
			//fmt.Sprintf("\r%s", str)
			if size >= int64(fileInfo.Length) && len(activeTaskList[fileInfo.Id]) <= 0 {
				fmt.Println("\n" + time.Now().String() + "下载完成")
				close(fileInfo.Exit)
			}
		}
	}
}

func allocation(fileId int, index int) {
	segList := main.readSeg(fileId)
	segment := segList[index]
	fileInfo := main.readFileMap(fileId)
	// 如果文件支持断点下载，且大于块大小
	//if fileMap[fileId].Renewal &&
	if segment.End-segment.Start > main.segSize {
		// 分块
		segment.End = segment.Start + main.segSize
		segNext := Downloader.SegMent{
			Start:    segment.End + 1,
			End:      fileInfo.Length - 1,
			Url:      fileInfo.Url,
			Count:    0,
			Index:    segment.Index + 1,
			Complete: false,
		}
		segList[index] = segment
		segList = append(segList, segNext)
		main.updateSeg(fileId, segList)
		allocation(fileId, index+1)
		return
	}
	return
}

func getPercent(a int64, b int64) int {
	result := float64(a) / float64(b) * 100
	return int(result)
}

func getCurrentSize(t time.Time) float64 {
	return time.Now().Sub(t).Seconds()
}
