package models

import "os"

type FileInfo struct {
	Id       int
	Row      int
	Renewal  bool // 是否支持断点续传
	Status   int
	Length   int
	Url      string
	File     *os.File
	FileName string
	FilePath string
	Exit     chan bool
	FileChan chan int
}
