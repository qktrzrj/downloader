package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"sync"
	"time"
	"yan.com/downloader/models"
)

const segSize int = 1024 * 1024

var (
	seg    = make(map[string][]models.SegMent)
	client fasthttp.Client
	group  sync.WaitGroup
)

func download(url string) error {
	// 获取目标文件信息
	fileInfo, err := getFileInfo(url)
	if err != nil {
		return fmt.Errorf("获取目标文件信息失败: %w", err)
	}

	group.Add(1)
	go downloadDirect(fileInfo)
	getRate(fileInfo, time.Now())
	group.Wait()
	return nil

}

func downloadDirect(fileInfo models.FileInfo) {
	segment := models.SegMent{
		Start:    0,
		End:      fileInfo.Length - 1,
		Url:      fileInfo.Url,
		Count:    0,
		Index:    0,
		Complete: false,
	}
	// 如果文件不支持断点续传，将不进行下载重试
	if !fileInfo.Renewal {
		segment.Count = 2
	}
	segList := make([]models.SegMent, 0)
	seg[fileInfo.FilePath] = append(segList, segment)
	go startBT(fileInfo, 0)
	// 开启任务

}
