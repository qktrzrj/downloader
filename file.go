package main

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"yan.com/downloader/models"
)

func isNotExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}
	return false
}

func creatFile(filePath string) (file *os.File, err error) {
	count := 2
	for {
		if !isNotExist(filePath) {
			// 重命名
			index := strings.LastIndex(filePath, ".")
			if index == -1 {
				filePath = filePath + strconv.Itoa(count)
				count++
				continue
			} else {
				temp := []byte(filePath)
				filePath = string(temp[0:index]) + strconv.Itoa(count) + string(temp[index:])
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

func writeFile(fileInfo models.FileInfo, segment *models.SegMent, body []byte) error {
	// 加锁
	fileInfo.Lock.Lock()
	// 偏移
	ret, err := fileInfo.File.Seek(int64(segment.Start), 0)
	if err != nil {
		return err
	}
	if ret != int64(segment.Start) {
		return errors.New("写入文件失败!")
	}
	// 写操作
	len, err := fileInfo.File.Write(body)
	if err != nil {

		return err
	}
	if len != segment.End-segment.Start {
		return errors.New("写入文件失败!")
	}
	fileInfo.Lock.Unlock()
	segment.Complete = true
	return nil
}
