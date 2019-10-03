package main

import (
	"downloader/models"
	"fmt"
	"time"
)

var schedule = make(chan int)
var sch = make(chan struct{})
var filenum int

func scheduler() {
	go func() {
		for {
			select {
			case <-sch:
				if filenum < 2 && fileQueue.Size() > 0 {
					fileId := *fileQueue.Dequeue()
					schedule <- fileId.(int)
					filenum++
				}
			}
		}
	}()
	for {
		select {
		case fileId := <-schedule:
			if filenum >= 3 {
				fileQueue.Enqueue(fileId)
				sch <- struct{}{}
				continue
			}
			go download(fileId)
		}
	}
}

func getRate(fileId int, t time.Time) {
	progressBar := barList[fileId]
	for {
		select {
		case <-fileMap[fileId].Exit:
			group.Done()
			return
		default:
			// 获取已写入块总和大小
			size := getFileSize(fileId)
			percent := getPercent(size, int64(fileMap[fileId].Length))
			//result, _ := strconv.Atoi(percent)
			//str := "working " + percent + "%" + "[" + bar(result, 100) + "] " + " " +
			//	fmt.Sprintf("%.f", getCurrentSize(t)) + "s"
			//fmt.Sprintf("\r%s", str)
			progressBar.SetValue(percent)
			if size >= int64(fileMap[fileId].Length) && len(activeTaskList[fileMap[fileId].Id]) <= 0 {
				fmt.Println("\n" + time.Now().String() + "下载完成")
				close(fileMap[fileId].Exit)
			}
		}
	}
}

func allocation(fileId int, index int) {
	segment := seg[fileId][index]
	// 如果文件支持断点下载，且大于块大小
	//if fileMap[fileId].Renewal &&
	if segment.End-segment.Start > segSize {
		// 分块
		segment.End = segment.Start + segSize
		segList := seg[fileId]
		segNext := models.SegMent{
			Start:    segment.End + 1,
			End:      fileMap[fileId].Length - 1,
			Url:      fileMap[fileId].Url,
			Count:    0,
			Index:    segment.Index + 1,
			Complete: false,
		}
		seg[fileId][index] = segment
		seg[fileId] = append(segList, segNext)
		allocation(fileId, index+1)
		return
	}
	return
}

func getPercent(a int64, b int64) int {
	result := a / b * 100
	return int(result)
}

func bar(count, size int) string {
	str := ""
	for i := 0; i < size; i++ {
		if i < count {
			str += "="
		} else {
			str += " "
		}
	}
	return str
}

func getCurrentSize(t time.Time) float64 {
	return time.Now().Sub(t).Seconds()
}
