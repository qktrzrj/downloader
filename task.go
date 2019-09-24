package main

import (
	"sync"
	"yan.com/downloader/models"
)

const (
	maxTaskSize int = 15
)

var (
	//taskOver = make([]task, 0)
	taskMap = make(map[string][]task)
	//activeTaskMap = make(map[string]int)
)

type task struct {
	segIndex int
	active   bool
	cache    []byte
}

func addTask(filePath string, segment models.SegMent) {
	index := segment.Index
	newTask := task{
		segIndex: index,
		active:   false,
		cache:    nil,
	}
	taskList, _ := getTaskList(filePath)
	taskList = append(taskList, newTask)
	setTaskList(filePath, taskList)
}

func startTask(filePath string) {
	//activeTaskMap[filePath] = 0
	taskGroup := sync.WaitGroup{}
	taskSize := len(seg[filePath])
	if taskSize > maxTaskSize {
		taskSize = maxTaskSize
	}
	exit := make(chan bool, taskSize)
	for i := 0; i < taskSize; i++ {
		taskGroup.Add(1)
		go func(exit <-chan bool) {
			for {
				select {
				case <-exit:
					taskGroup.Done()
					return
				default:
					taskList, ok := getTaskList(filePath)
					if !ok {
						continue
					}
					if len(taskList) <= 0 {
						continue
					}
					taskInfo := taskList[0]
					taskList = taskList[1:]
					setTaskList(filePath, taskList)
					segment := &seg[filePath][taskInfo.segIndex]
					request := getRequest(segment.Url, segment.Start, segment.End)
					send(request, segment, fileMap[filePath])
					go writeFile(*fileMap[filePath], segment)
				}
			}
		}(exit)
	}
	select {
	case <-fileMap[filePath].TaskExit:
		for i := 0; i < taskSize; i++ {
			exit <- true
		}
		taskGroup.Wait()
		fileMap[filePath].RateExit <- true
		return
	}
}

func getTaskList(filePath string) ([]task, bool) {
	fileMap[filePath].Task.RLock()
	taskList, ok := taskMap[filePath]
	fileMap[filePath].Task.RUnlock()
	return taskList, ok
}

func setTaskList(filePath string, taskList []task) {
	fileMap[filePath].Task.Lock()
	taskMap[filePath] = taskList
	fileMap[filePath].Task.Unlock()
}
