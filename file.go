package main

import (
	"os"
	"strconv"
	"strings"
	"yan.com/downloader/models"
)

var fileMap = make(map[string]*models.FileInfo)

func isNotExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}
	return false
}

func creatFile(filePath string) (file *os.File, err error) {
	fileOrgin := filePath
	temp := []byte(fileOrgin)
	// 重命名
	index := strings.LastIndex(filePath, ".")
	pre := string(temp[0:index])
	end := string(temp[index:])
	count := 2
	for {
		if !isNotExist(filePath) {
			if index == -1 {
				filePath = fileOrgin + strconv.Itoa(count)
				count++
				continue
			} else {
				filePath = pre + strconv.Itoa(count) + end
				count++
				continue
			}
		}
		break
	}
	file, err = os.Create(filePath)
	//file, err = os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	return
}

func deleteFile(filePath string) {
	if isNotExist(filePath) {
		return
	}
	os.Remove(filePath)
}

func getFileSize(filePath string) int64 {
	if isNotExist(filePath) {
		return 0
	}
	fi, _ := os.Stat(filePath)
	return fi.Size()
}

func writeFile(fileInfo models.FileInfo, segment *models.SegMent) {
	// 加锁
	//fileInfo.Lock.Lock()
	file := fileInfo.File
	// 写操作
	len, err := file.WriteAt(segment.Cache, int64(segment.Start))
	if err != nil {
		//fileInfo.Lock.Unlock()
		return
	}
	if len-1 != segment.End-segment.Start {
		//fileInfo.Lock.Unlock()
		return
	}
	//fileInfo.Lock.Unlock()
}
