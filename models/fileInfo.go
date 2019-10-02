package models

import "os"

type FileInfo struct {
	Id       int
	Renewal  bool // 是否支持断点续传
	Length   int
	Url      string
	File     *os.File
	FileName string
	FilePath string
	Exit     chan bool
	FileChan chan int
}
