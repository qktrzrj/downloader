package main

import (
	"downloader/models"
	"os"
	"strconv"
	"strings"
)

var fileMap = make(map[int]models.FileInfo)
var fileQueue ItemQueue

func isNotExist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}
	return false
}

func creatFile(filePath string) (file *os.File, err error) {
	fileOrigin := filePath
	temp := []byte(fileOrigin)
	// 重命名
	index := strings.LastIndex(filePath, ".")
	pre := string(temp[0:index])
	end := string(temp[index:])
	count := 2
	for {
		if !isNotExist(filePath) {
			if index == -1 {
				filePath = fileOrigin + strconv.Itoa(count)
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
	return
}

func deleteFile(filePath string) {
	if isNotExist(filePath) {
		return
	}
	os.Remove(filePath)
}

func getFileSize(fileId int) int64 {
	if isNotExist(fileMap[fileId].FilePath) {
		return 0
	}
	fi, _ := fileMap[fileId].File.Stat()
	return fi.Size()
}

func writeFile(fileId int) {
	for {
		select {
		case <-fileMap[fileId].Exit:
			return
		default:
			index := <-fileMap[fileId].FileChan
			segment := seg[fileId][index]
			file := fileMap[fileId].File
			// 写操作
			len, err := file.WriteAt(segment.Cache, int64(segment.Start))
			file.Sync()
			if err != nil {
				fileMap[fileId].FileChan <- index
			}
			if len-1 != segment.End-segment.Start {
				fileMap[fileId].FileChan <- index
			}
			segment.Complete = true
			segment.Cache = nil
			seg[fileId][segment.Index] = segment
		}
	}
}
