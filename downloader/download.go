package downloader

import (
	"downloader"
	"sync"
	"time"
)

var (
	segLock     = sync.RWMutex{}
	seg         = make(map[int][]Downloader.SegMent)
	segSize int = 0.5 * 1024 * 1024
	barList     = make(map[int]int)
	group   sync.WaitGroup
)

func updateSeg(fileId int, segList []Downloader.SegMent) {
	segLock.Lock()
	seg[fileId] = segList
	segLock.Unlock()
}

func updateSegIndex(fileId int, index int, segment Downloader.SegMent) {
	segLock.Lock()
	seg[fileId][index] = segment
	segLock.Unlock()
}

func readSeg(fileId int) (segList []Downloader.SegMent) {
	segLock.RLock()
	segList = seg[fileId]
	segLock.RUnlock()
	return
}

func readSegIndex(fileId int, index int) (segment Downloader.SegMent) {
	segLock.RLock()
	segment = seg[fileId][index]
	segLock.RUnlock()
	return
}

func download(fileId int) {
	group.Add(1)
	// 开启任务
	go startBT(fileId)
	go startTask(fileId)
	go main.writeFile(fileId)
	go getRate(fileId, time.Now())
	group.Wait()
	filenum--
	sch <- struct{}{}
	time.Sleep(time.Second)
	main.fileMap[fileId].File.Close()
	delete(seg, fileId)
	delete(taskList, fileId)
	return
}

func downloadDirect(fileId int) {
	segment := Downloader.SegMent{
		Start:    0,
		End:      main.fileMap[fileId].Length - 1,
		Url:      main.fileMap[fileId].Url,
		Count:    0,
		Index:    0,
		Complete: false,
	}
	// 如果文件不支持断点续传，将不进行下载重试
	//if !fileMap[fileId].Renewal {
	//	segment.Count = 2
	//}
	segList := []Downloader.SegMent{segment}
	updateSeg(fileId, segList)
	// 分配块
	Downloader.allocation(fileId, 0)
	// 分配下载channel
	info := main.readFileMap(fileId)
	info.FileChan = make(chan int, len(seg[fileId]))
	main.updateFileMap(fileId, info)
}
