package main

import (
	"fmt"
	"strconv"
	"time"
	"yan.com/downloader/models"
)

func getRate(fileInfo models.FileInfo, time time.Time) {
	for {
		// 获取文件大小
		size := getFileSize(fileInfo.FilePath)
		percent := getPercent(size, int64(fileInfo.Length))
		result, _ := strconv.Atoi(percent)
		str := "working " + percent + "%" + "[" + bar(result, 100) + "] " + " " +
			fmt.Sprintf("%.f", getCurrentSize(time)) + "s"
		fmt.Printf("\r%s", str)
		if size == int64(fileInfo.Length) {
			fmt.Println("\n下载完成")
			fileInfo.File.Close()
			group.Done()
			break
		}
	}
}

func startBT(fileInfo models.FileInfo, index int) {
	segment := &seg[fileInfo.FilePath][index]
	// 如果文件支持断点下载，且大于块大小
	if fileInfo.Renewal && segment.End-segment.Start > segSize {
		// 分块
		segment.End = segment.Start + segSize
		segList := seg[fileInfo.FilePath]
		segNext := models.SegMent{
			Start:    segment.End + 1,
			End:      fileInfo.Length - 1,
			Url:      fileInfo.Url,
			Count:    0,
			Index:    segment.Index + 1,
			Complete: false,
		}
		seg[fileInfo.FilePath] = append(segList, segNext)
		// 加入队列
		addTask(fileInfo.FilePath, *segment)
		startBT(fileInfo, index+1)
	}
}

func getPercent(a int64, b int64) string {
	result := float64(a) / float64(b) * 100
	return fmt.Sprintf("%.f", result)
}

func bar(count, size int) string {
	str := ""
	for i := 0; i < size; i++ {
		if i < count {
			str += "="
		} else {
			str += " "
		}
	}
	return str
}

func getCurrentSize(t time.Time) float64 {
	return time.Now().Sub(t).Seconds()
}
