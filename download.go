package main

import (
	"downloader/models"
	"github.com/andlabs/ui"
	"sync"
	"time"
)

const segSize int = 0.5 * 1024 * 1024

var (
	seg     = make(map[int][]models.SegMent)
	barList = make(map[int]*ui.ProgressBar)
	group   sync.WaitGroup
)

func download(fileId int) {
	group.Add(1)
	downloadDirect(fileId)
	// 开启任务
	go startBT(fileId)
	go startTask(fileId)
	go writeFile(fileId)
	//downloader.Show()
	go getRate(fileId, time.Now())
	group.Wait()
	filenum--
	sch <- struct{}{}
	fileMap[fileId].File.Close()
	delete(seg, fileId)
	delete(taskList, fileId)
	return
}

func downloadDirect(fileId int) {
	segment := models.SegMent{
		Start:    0,
		End:      fileMap[fileId].Length - 1,
		Url:      fileMap[fileId].Url,
		Count:    0,
		Index:    0,
		Complete: false,
	}
	// 如果文件不支持断点续传，将不进行下载重试
	//if !fileMap[fileId].Renewal {
	//	segment.Count = 2
	//}
	segList := make([]models.SegMent, 0)
	seg[fileMap[fileId].Id] = append(segList, segment)
	// 分配块
	allocation(fileId, 0)
	// 分配下载channel
	info := fileMap[fileId]
	info.FileChan = make(chan int, len(seg[fileId]))
	fileMap[fileId] = info
}
