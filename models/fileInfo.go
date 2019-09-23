package models

import (
	"os"
	"sync"
)

type FileInfo struct {
	Renewal  bool // 是否支持断点续传
	Length   int
	Url      string
	FileName string
	FilePath string
	File     *os.File
	Exit     chan bool
	Lock     sync.Mutex
	Down     bool
}
