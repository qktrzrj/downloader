package models

import (
	"os"
	"sync"
)

type FileInfo struct {
	Renewal   bool // 是否支持断点续传
	Length    int
	Url       string
	FileName  string
	FilePath  string
	File      *os.File
	Exit      chan bool
	TaskExit  chan bool
	RateExit  chan bool
	Lock      sync.Mutex   //文件锁
	Task      sync.RWMutex //任务队列锁
	Down      bool
	FileCache map[int][]byte
}
