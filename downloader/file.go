package downloader

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	fileMap   = make(map[int]Downloader.task)
	fileLock  = sync.RWMutex{}
	fileQueue ItemQueue
)

func updateFileMap(fileId int, fileInfo Downloader.task) {
	fileLock.Lock()
	fileMap[fileId] = fileInfo
	fileLock.Unlock()
}

func readFileMap(fileId int) (fileInfo Downloader.task) {
	fileLock.RLock()
	fileInfo = fileMap[fileId]
	fileLock.RUnlock()
	return
}

func FileIsNotExist(filePath string) bool {
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
	fileInfo := readFileMap(fileId)
	if isNotExist(fileInfo.FilePath) {
		return 0
	}
	fi, _ := fileInfo.File.Stat()
	return fi.Size()
}

func writeFile(fileId int) {
	for {
		fileInfo := readFileMap(fileId)
		select {
		case <-fileInfo.Exit:
			fmt.Println("退出写文件")
		case <-fileInfo.Pause:
			fmt.Println("暂停写文件")
			select {
			case <-fileInfo.Continue:
				fmt.Println("继续写文件")
			}
		default:
			index := <-fileInfo.FileChan
			fmt.Println(index)
			segList := readSeg(fileId)
			segment := segList[index]
			// 写操作
			len, err := fileInfo.File.WriteAt(segment.Cache, int64(segment.Start))
			if err != nil {
				fileInfo.FileChan <- index
			}
			if len-1 != segment.End-segment.Start {
				fileInfo.FileChan <- index
			}
			segment.Complete = true
			segment.Cache = nil
			updateSegIndex(fileId, index, segment)
		}
	}
}
