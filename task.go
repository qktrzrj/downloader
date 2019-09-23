package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"yan.com/downloader/models"
)

const (
	maxSize     int = 3
	maxTaskSize int = 10
)

var (
	taskMap       = make(map[string][]task)
	activeTaskMap = make(map[string][maxTaskSize]int)
)

type task struct {
	filePath string
	segIndex int
	active   bool
	cache    []byte
}

func addTask(filePath string, segment models.SegMent) {
	newTask := task{
		filePath: filePath,
		segIndex: segment.Index,
		active:   false,
		cache:    nil,
	}
	taskList := taskMap[filePath]
	taskList = append(taskList, newTask)
	taskMap[filePath] = taskList
}

func initTask() {
	count := 0
	for index, fileInfo := range fileList {
		activeTaskList := activeTaskMap[fileInfo.FilePath]
		if count >= maxTaskSize {
			break
		}
		activeTaskList[count] = index
		count++
	}
}

func startTask() {
	initTask()
	for {
		for filePath, taskInfo := range activeTaskMap {
			request := getRequest(fileInfo.Url, segment.Start, segment.End)
			resp := &fasthttp.Response{}
			ok := make(chan bool)
			go send(request, resp, segment, fileInfo, &ok)
			for {
				select {
				case <-fileInfo.Down:
					fmt.Println("文件下载失败")
					return
				case <-ok:
					fmt.Println("块下载成功")
					return
				}
			}
		}

	}

}
